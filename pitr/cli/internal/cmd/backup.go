/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/apache/shardingsphere-on-cloud/pitr/cli/internal/pkg"
	"github.com/apache/shardingsphere-on-cloud/pitr/cli/internal/pkg/model"
	"github.com/apache/shardingsphere-on-cloud/pitr/cli/internal/pkg/xerr"
	"github.com/apache/shardingsphere-on-cloud/pitr/cli/pkg/logging"
	"github.com/apache/shardingsphere-on-cloud/pitr/cli/pkg/prettyoutput"
	"github.com/apache/shardingsphere-on-cloud/pitr/cli/pkg/promptutil"
	"github.com/apache/shardingsphere-on-cloud/pitr/cli/pkg/timeutil"

	"github.com/google/uuid"
	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
)

const (
	// defaultInstance is used to set backup instance name in openGauss, we can modify it in the future.
	defaultInstance = "ins-default-ss"
	// defaultShowDetailRetryTimes retry times of check backup detail from agent server
	defaultShowDetailRetryTimes = 3

	backupPromptFmt = "Please Check All Nodes Disk Space, Make Sure Have Enough Space To Backup Or Restore Data.\n" +
		"Are you sure to continue? (Y/N)"
)

var filename string

var BackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup a database cluster",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			fmt.Printf("Flag: %s Value: %s\n", flag.Name, flag.Value)
		})

		// convert BackupModeStr to BackupMode
		switch BackupModeStr {
		case "FULL", "full":
			BackupMode = model.DBBackModeFull
		case "PTRACK", "ptrack":
			BackupMode = model.DBBackModePTrack
		}
		if BackupMode == model.DBBackModePTrack {
			logging.Warn("Please make sure all openGauss nodes have been set correct configuration about ptrack. You can refer to https://support.huaweicloud.com/intl/zh-cn/devg-opengauss/opengauss_devg_1362.html for more details.")
		}

		logging.Info(fmt.Sprintf("Default backup path: %s", pkg.DefaultRootDir()))

		// Start backup
		if err := backup(); err != nil {
			logging.Error(err.Error())
		}
	},
}

func init() {
	RootCmd.AddCommand(BackupCmd)

	BackupCmd.Flags().StringVarP(&Host, "host", "H", "", "ss-proxy hostname or ip")
	_ = BackupCmd.MarkFlagRequired("host")
	BackupCmd.Flags().Uint16VarP(&Port, "port", "P", 0, "ss-proxy port")
	_ = BackupCmd.MarkFlagRequired("port")
	BackupCmd.Flags().StringVarP(&Username, "username", "u", "", "ss-proxy username")
	_ = BackupCmd.MarkFlagRequired("username")
	BackupCmd.Flags().StringVarP(&Password, "password", "p", "", "ss-proxy password")
	_ = BackupCmd.MarkFlagRequired("password")
	BackupCmd.Flags().StringVarP(&BackupPath, "dn-backup-path", "B", "", "openGauss data backup path")
	_ = BackupCmd.MarkFlagRequired("dn-backup-path")
	BackupCmd.Flags().StringVarP(&BackupModeStr, "dn-backup-mode", "b", "", "openGauss data backup mode (FULL|PTRACK)")
	_ = BackupCmd.MarkFlagRequired("dn-backup-mode")
	BackupCmd.Flags().Uint8VarP(&ThreadsNum, "dn-threads-num", "j", 1, "openGauss data backup threads nums")
	BackupCmd.Flags().Uint16VarP(&AgentPort, "agent-port", "a", 443, "agent server port")
	_ = BackupCmd.MarkFlagRequired("agent-port")

}

// Steps of backup:
// 1. lock cluster
// 2. Get cluster info and save local backup info
// 3. Operate backup by agent-server
// 4. unlock cluster
// 5. Waiting for backups finished
// 6. Update local backup info
// 7. Double check backups all finished
// nolint:gocognit
func backup() error {
	var (
		err      error
		lsBackup *model.LsBackup
		cancel   bool
	)
	proxy, err := pkg.NewShardingSphereProxy(Username, Password, pkg.DefaultDBName, Host, Port)
	if err != nil {
		return xerr.NewCliErr(fmt.Sprintf("Connect shardingsphere proxy failed, err: %s", err))
	}

	ls, err := pkg.NewLocalStorage(pkg.DefaultRootDir())
	if err != nil {
		return xerr.NewCliErr(fmt.Sprintf("Create local storage failed. err: %s", err))
	}

	defer func() {
		if err != nil {
			if !cancel {
				logging.Warn("Try to unlock cluster ...")
			}
			if err := proxy.Unlock(); err != nil {
				logging.Error(fmt.Sprintf("Since backup failed, try to unlock cluster, but still failed. err: %s", err))
			}

			if lsBackup != nil {
				if cancel {
					deleteBackupFiles(ls, lsBackup, deleteModeQuiet)
				} else {
					logging.Warn("Try to delete backup data ...")
					deleteBackupFiles(ls, lsBackup, deleteModeNormal)
				}
			}
		}
	}()

	// Step1. lock cluster
	logging.Info("Starting lock cluster ...")
	err = proxy.LockForBackup()
	if err != nil {
		return xerr.NewCliErr(fmt.Sprintf("Lock for backup failed. err: %s", err))
	}

	// Step2. Get cluster info and save local backup info
	logging.Info("Starting export metadata ...")
	lsBackup, err = exportData(proxy, ls)
	if err != nil {
		return xerr.NewCliErr(fmt.Sprintf("export backup data failed. err: %s", err))
	}
	logging.Info(fmt.Sprintf("Export backup data success, backup filename: %s", filename))

	// Step3. Check agent server status
	logging.Info("Checking agent server status...")
	if available := checkAgentServerStatus(lsBackup); !available {
		logging.Error("Cancel! One or more agent server are not available.")
		err = xerr.NewCliErr("One or more agent server are not available.")
		return err
	}

	// Step4. Show disk space
	logging.Info("Checking disk space...")
	err = checkDiskSpace(lsBackup)
	if err != nil {
		return xerr.NewCliErr(fmt.Sprintf("check disk space failed. err: %s", err))
	}

	prompt := fmt.Sprintln(backupPromptFmt)
	err = promptutil.GetUserApproveInTerminal(prompt)
	if err != nil {
		cancel = true
		return xerr.NewCliErr(fmt.Sprintf("%s", err))
	}

	// Step5. send backup command to agent-server.
	logging.Info("Starting backup ...")
	err = execBackup(lsBackup)
	if err != nil {
		return xerr.NewCliErr(fmt.Sprintf("exec backup failed. err: %s", err))
	}

	// Step6. unlock cluster
	logging.Info("Starting unlock cluster ...")
	err = proxy.Unlock()
	if err != nil {
		return xerr.NewCliErr(fmt.Sprintf("unlock cluster failed. err: %s", err))
	}

	// Step7. update backup file
	logging.Info("Starting update backup file ...")
	err = ls.WriteByJSON(filename, lsBackup)
	if err != nil {
		return xerr.NewCliErr(fmt.Sprintf("update backup file failed. err: %s", err))
	}

	// Step8. check agent server backup
	logging.Info("Starting check backup status ...")
	status := checkBackupStatus(lsBackup)
	logging.Info(fmt.Sprintf("Backup result: %s", status))
	if status != model.SsBackupStatusCompleted && status != model.SsBackupStatusCanceled {
		err = xerr.NewCliErr("Backup failed")
		return err
	}

	// Step9. finished backup and update backup file
	logging.Info("Starting update backup file ...")
	err = ls.WriteByJSON(filename, lsBackup)
	if err != nil {
		return xerr.NewCliErr(fmt.Sprintf("update backup file failed. err: %s", err))
	}

	logging.Info("Backup finished!")
	return nil
}

func exportData(proxy pkg.IShardingSphereProxy, ls pkg.ILocalStorage) (lsBackup *model.LsBackup, err error) {
	// Step1. export cluster metadata from ss-proxy
	cluster, err := proxy.ExportMetaData()
	if err != nil {
		return nil, xerr.NewCliErr(fmt.Sprintf("export meta data failed. err: %s", err))
	}

	// Step2. export storage nodes from ss-proxy
	nodes, err := proxy.ExportStorageNodes()
	if err != nil {
		return nil, xerr.NewCliErr(fmt.Sprintf("export storage nodes failed. err: %s", err))
	}

	// Step3. combine the backup contents
	filename = ls.GenFilename(pkg.ExtnJSON)
	csn := ""
	if cluster.SnapshotInfo != nil {
		csn = cluster.SnapshotInfo.Csn
	}

	contents := &model.LsBackup{
		Info: &model.BackupMetaInfo{
			ID:         uuid.New().String(), // generate uuid for this backup
			CSN:        csn,
			StartTime:  timeutil.Now().String(),
			EndTime:    timeutil.Init(),
			BackupMode: BackupMode,
		},
		SsBackup: &model.SsBackup{
			Status:       model.SsBackupStatusWaiting, // default status of backup is model.SsBackupStatusWaiting
			ClusterInfo:  cluster,
			StorageNodes: nodes,
		},
	}

	// Step4. finally, save data with json to local
	if err := ls.WriteByJSON(filename, contents); err != nil {
		return nil, xerr.NewCliErr(fmt.Sprintf("write backup info by json failed. err: %s", err))
	}

	return contents, nil
}

func execBackup(lsBackup *model.LsBackup) error {
	sNodes := lsBackup.SsBackup.StorageNodes
	dnCh := make(chan *model.DataNode, len(sNodes))
	g := new(errgroup.Group)

	logging.Info("Starting send backup command to agent server...")

	for _, node := range sNodes {
		sn := node
		as := pkg.NewAgentServer(fmt.Sprintf("%s:%d", convertLocalhost(sn.IP), AgentPort))
		g.Go(func() error {
			return _execBackup(as, sn, dnCh)
		})
	}

	err := g.Wait()
	close(dnCh)

	// if backup failed, return error
	if err != nil {
		lsBackup.SsBackup.Status = model.SsBackupStatusFailed
		return xerr.NewCliErr(fmt.Sprintf("node backup failed. err: %s", err))
	}

	// save data node list to lsBackup
	for dn := range dnCh {
		lsBackup.DnList = append(lsBackup.DnList, dn)
	}

	lsBackup.SsBackup.Status = model.SsBackupStatusRunning
	return nil
}

func _execBackup(as pkg.IAgentServer, node *model.StorageNode, dnCh chan *model.DataNode) error {
	in := &model.BackupIn{
		DBPort:       node.Port,
		DBName:       node.Database,
		Username:     node.Username,
		Password:     node.Password,
		DnBackupPath: BackupPath,
		DnThreadsNum: ThreadsNum,
		DnBackupMode: BackupMode,
		Instance:     defaultInstance,
	}
	backupID, err := as.Backup(in)
	if err != nil {
		return xerr.NewCliErr(fmt.Sprintf("backup failed, err: %s", err))
	}

	// update DnList of lsBackup
	dn := &model.DataNode{
		IP:        node.IP,
		Port:      node.Port,
		Status:    model.SsBackupStatusRunning,
		BackupID:  backupID,
		StartTime: timeutil.Now().String(),
		EndTime:   timeutil.Init(),
	}
	dnCh <- dn
	return nil
}

func checkBackupStatus(lsBackup *model.LsBackup) model.BackupStatus {
	var (
		dataNodeMap       = make(map[string]*model.DataNode)
		dnCh              = make(chan *model.DataNode, len(lsBackup.DnList))
		backupFinalStatus = model.SsBackupStatusCompleted
		totalNum          = len(lsBackup.SsBackup.StorageNodes)
		dnResult          = make([]*model.DataNode, 0)
	)

	if totalNum == 0 {
		logging.Info("No data node need to backup")
		return model.SsBackupStatusCanceled
	}

	for _, dn := range lsBackup.DnList {
		dataNodeMap[dn.IP] = dn
	}

	pw := prettyoutput.NewProgressPrinter(prettyoutput.ProgressPrintOption{
		NumTrackersExpected: totalNum,
	})

	go pw.Render()

	for i := 0; i < totalNum; i++ {
		sn := lsBackup.SsBackup.StorageNodes[i]
		as := pkg.NewAgentServer(fmt.Sprintf("%s:%d", convertLocalhost(sn.IP), AgentPort))
		dn := dataNodeMap[sn.IP]
		backupInfo := &model.BackupInfo{}
		task := &backuptask{
			As:      as,
			Sn:      sn,
			Dn:      dn,
			DnCh:    dnCh,
			Backup:  backupInfo,
			retries: defaultShowDetailRetryTimes,
		}
		tracker := &progress.Tracker{
			Message: fmt.Sprintf("Checking backup status  # %s:%d", sn.IP, sn.Port),
			Total:   0,
			Units:   progress.UnitsDefault,
		}
		pw.AppendTracker(tracker)
		go pw.UpdateProgress(tracker, task.checkProgress)
	}

	pw.BlockedRendered()

	close(dnCh)

	for dn := range dnCh {
		dnResult = append(dnResult, dn)
		if dn.Status != model.SsBackupStatusCompleted {
			backupFinalStatus = model.SsBackupStatusFailed
		}
	}

	// print backup result formatted
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetTitle("Backup Task Result: %s", backupFinalStatus)
	t.AppendHeader(table.Row{"#", "Data Node IP", "Data Node Port", "Result"})

	for i, dn := range dnResult {
		t.AppendRow([]interface{}{i + 1, dn.IP, dn.Port, dn.Status})
		t.AppendSeparator()
	}

	t.Render()

	lsBackup.DnList = dnResult
	lsBackup.SsBackup.Status = backupFinalStatus
	lsBackup.Info.EndTime = timeutil.Now().String()
	return backupFinalStatus
}

type backuptask struct {
	As   pkg.IAgentServer
	Sn   *model.StorageNode
	Dn   *model.DataNode
	DnCh chan *model.DataNode

	Backup  *model.BackupInfo
	retries int
}

func (t *backuptask) checkProgress() (bool, error) {
	var err error
	in := &model.ShowDetailIn{
		DBPort:       t.Sn.Port,
		DBName:       t.Sn.Database,
		Username:     t.Sn.Username,
		Password:     t.Sn.Password,
		DnBackupID:   t.Dn.BackupID,
		DnBackupPath: BackupPath,
		Instance:     defaultInstance,
	}

	t.Backup, err = t.As.ShowDetail(in)
	if err != nil {
		if t.retries == 0 {
			t.Dn.Status = model.SsBackupStatusCheckError
			t.DnCh <- t.Dn
			return false, err
		}
		time.Sleep(time.Second * 1)
		t.retries--
		return t.checkProgress()
	}

	t.Dn.Status = t.Backup.Status
	t.Dn.EndTime = timeutil.Now().String()

	if t.Backup.Status == model.SsBackupStatusCompleted || t.Backup.Status == model.SsBackupStatusFailed {
		t.DnCh <- t.Dn
		return true, nil
	}
	return false, nil
}

type deleteMode int

const (
	deleteModeNormal deleteMode = iota
	deleteModeQuiet
)

func deleteBackupFiles(ls pkg.ILocalStorage, lsBackup *model.LsBackup, m deleteMode) {
	var (
		dataNodeMap = make(map[string]*model.DataNode)
		totalNum    = len(lsBackup.SsBackup.StorageNodes)
		resultCh    = make(chan *model.DeleteBackupResult, totalNum)
	)
	for _, dn := range lsBackup.DnList {
		dataNodeMap[dn.IP] = dn
	}

	if totalNum == 0 {
		logging.Info("No data node need to delete backup files")
		return
	}

	pw := prettyoutput.NewPW(totalNum)
	go pw.Render()

	for _, sn := range lsBackup.SsBackup.StorageNodes {
		sn := sn
		dn, ok := dataNodeMap[sn.IP]
		if !ok {
			if m != deleteModeQuiet {
				logging.Warn(fmt.Sprintf("SKIPPED! data node %s:%d not found in backup info.", sn.IP, sn.Port))
			}
			continue
		}
		as := pkg.NewAgentServer(fmt.Sprintf("%s:%d", convertLocalhost(sn.IP), AgentPort))

		go doDelete(as, sn, dn, resultCh, pw)
	}

	if m != deleteModeQuiet {

		time.Sleep(time.Millisecond * 100)
		for pw.IsRenderInProgress() {
			if pw.LengthActive() == 0 {
				pw.Stop()
			}
			time.Sleep(time.Millisecond * 100)
		}

		close(resultCh)

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetTitle("Delete Backup Files Result")
		t.AppendHeader(table.Row{"#", "Node IP", "Node Port", "Result", "Message"})
		t.SetColumnConfigs([]table.ColumnConfig{{Number: 5, WidthMax: 50}})

		idx := 0
		for result := range resultCh {
			idx++
			t.AppendRow([]interface{}{idx, result.IP, result.Port, result.Status, result.Msg})
			t.AppendSeparator()
		}

		t.Render()
	}

	if err := ls.DeleteByName(filename); err != nil {
		logging.Warn("Delete backup info file failed")
	}

	if m != deleteModeQuiet {
		logging.Info("Delete backup files finished")
	}
}

func doDelete(as pkg.IAgentServer, sn *model.StorageNode, dn *model.DataNode, resultCh chan *model.DeleteBackupResult, pw progress.Writer) {
	var (
		tracker = progress.Tracker{Message: fmt.Sprintf("Deleting backup files  # %s:%d", sn.IP, sn.Port), Total: 0, Units: progress.UnitsDefault}
	)

	pw.AppendTracker(&tracker)

	in := &model.DeleteBackupIn{
		DBPort:       sn.Port,
		DBName:       sn.Database,
		Username:     sn.Username,
		Password:     sn.Password,
		DnBackupPath: BackupPath,
		BackupID:     dn.BackupID,
		Instance:     defaultInstance,
	}

	r := &model.DeleteBackupResult{
		IP:   sn.IP,
		Port: sn.Port,
	}

	if err := as.DeleteBackup(in); err != nil {
		r.Status = model.SsBackupStatusFailed
		r.Msg = err.Error()
		resultCh <- r
		tracker.MarkAsErrored()
	} else {
		tracker.MarkAsDone()
		r.Status = model.SsBackupStatusCompleted
		resultCh <- r
	}
}
