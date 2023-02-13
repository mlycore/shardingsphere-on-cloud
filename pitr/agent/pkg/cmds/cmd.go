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

package cmds

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"

	"github.com/apache/shardingsphere-on-cloud/pitr/agent/pkg/syncutils"
)

type Output struct {
	LineNo  uint32 // Start 1
	Message string
	Error   error
}

func Commands(name string, args ...string) (chan *Output, error) {
	c := "-c"
	args = append([]string{c}, args...)

	cmd := exec.Command(name, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("can not obtain stdout pipe for command[args=%+v]:%s", args, err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("the command is err[args=%+v]:%s", args, err)
	}

	reader := bufio.NewReader(stdout)

	output := make(chan *Output, 10)
	index := uint32(1)

	go func() {
		if err := syncutils.NewRecoverFuncWithErrRet("", func() error {
			for {
				msg, err := reader.ReadString('\n')
				if io.EOF == err {
					goto end
				} else if err != nil {
					output <- &Output{
						LineNo:  index,
						Message: msg,
						Error:   err,
					}
					goto end
				}

				output <- &Output{
					LineNo:  index,
					Message: msg,
				}

				index++
			}
		end:
			if err := cmd.Wait(); err != nil {
				output <- &Output{
					Error: err,
				}
			}

			return nil
		})(); err != nil {
			// only panic err
			output <- &Output{
				Error: err,
			}
		}
		close(output)
	}()

	return output, nil
}
