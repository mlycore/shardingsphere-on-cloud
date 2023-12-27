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

package pkg

import (
	"errors"
	"fmt"
	"strings"

	"github.com/apache/shardingsphere-on-cloud/pitr/agent/internal/cons"
	"github.com/apache/shardingsphere-on-cloud/pitr/agent/pkg/cmds"
	"github.com/apache/shardingsphere-on-cloud/pitr/agent/pkg/logging"
)

type (
	shardingsphere struct {
		shell string
		path  string
		log   logging.ILog
	}

	IShardingSphere interface {
		Restart(args []string) error
	}
)

var _ IShardingSphere = (*shardingsphere)(nil)

func NewShardingSphere(shell, path string, log logging.ILog) IShardingSphere {
	return &shardingsphere{
		shell: shell,
		path:  path,
		log:   log,
	}
}

const (
	_stopFmt  = "%s/bin/stop.sh"
	_startFmt = "%s/bin/start.sh %s"
)

func (s *shardingsphere) Restart(args []string) error {
	var argstr string
	if len(args) != 0 {
		argstr = strings.Join(args, " ")
	}

	stopcmd := fmt.Sprintf(_stopFmt, s.path)
	output, err := cmds.Exec(s.shell, stopcmd)
	s.log.Debug(fmt.Sprintf("stop shardingsphere output[msg=%s,err=%v]", output, err))

	if errors.Is(err, cons.CmdOperateFailed) {
		return fmt.Errorf("stop shardingsphere failure,err=%s,wrap=%w", err, cons.StopShardingSphereFailed)
	}
	if err != nil {
		return fmt.Errorf("cmds.Exec[shell=%s,cmd=%s] return err=%w", s.shell, stopcmd, err)
	}

	startcmd := fmt.Sprintf(_startFmt, s.path, argstr)
	output, err = cmds.Exec(s.shell, startcmd)
	s.log.Debug(fmt.Sprintf("start shardingsphere output[msg=%s,err=%v]", output, err))

	if errors.Is(err, cons.CmdOperateFailed) {
		return fmt.Errorf("start shardingsphere failure,err=%s,wrap=%w", err, cons.StartShardingSphereFailed)
	}
	if err != nil {
		return fmt.Errorf("cmds.Exec[shell=%s,cmd=%s] return err=%w", s.shell, startcmd, err)
	}

	return nil

}
