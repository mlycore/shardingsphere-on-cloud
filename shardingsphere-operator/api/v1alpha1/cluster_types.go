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

package v1alpha1

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

// +kubebuilder:object:root=true
// ComputeNode is the Schema for the ShardingSphere Proxy API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ClusterSpec `json:"spec,omitempty"`
	// +optional
	// Status ClusterStatus `json:"status,omitempty"`
}

type ClusterSpec struct {
	Encryption EncryptRule
	Topology   ClusterTopology
}

type ClusterTopology struct {
	ComputeNode ComputeNodeSpec
	StorageNode StorageNodeSpec
}

// type CreateEncryptRule struct {
// 	EncryptRule
// }

// func (t *CreateEncryptRule) ToDistSQL() string {
// 	return fmt.Sprintf("CREATE ENCRYPT RULE %s", t.EncryptRule.ToDistSQL())
// }

type EncryptRule struct {
	// ifNotExists
	// IfNotExist IfNotExists
	// IfNotExist bool
	// encryptDefinition
	// EncryptDefinitions EncryptDefinitionList
}

func (t *EncryptRule) ToDistSQL() string {
	var stmt string
	if t.IfNotExist {
		stmt = fmt.Sprintf("%s", t.IfNotExist.ToDistSQL())
	}
	stmt = fmt.Sprintf("%s %s;", stmt, t.EncryptDefinitions.ToDistSQL())

	return stmt
}

type EncryptDefinitionList []EncryptDefinition

func (t *EncryptDefinitionList) ToDistSQL() string {
	var stmt string
	for _, def := range *t {
		stmt = fmt.Sprintf("%s %s,", stmt, def.ToDistSQL())
	}
	stmt = strings.TrimSuffix(stmt, ",")
	return stmt
}

type IfNotExists bool

func (t IfNotExists) ToDistSQL() string {
	return "IF NOT EXISTS"
}

type EncryptDefinition struct {
	// name
	Name string
	// columns
	Columns ColumnList
	// queryWithCipheColumn
	QueryWithCipherColumn QueryWithCipherColumn
}

func (t *EncryptDefinition) ToDistSQL() string {
	var stmt string
	stmt = fmt.Sprintf("%s (COLUMNS %s), %s)", t.Name, t.Columns.ToDistSQL(), t.QueryWithCipherColumn.ToDistSQL())
	return stmt
}

type ColumnList []Column

func (t *ColumnList) ToDistSQL() string {
	var stmt string
	for _, c := range *t {
		stmt = fmt.Sprintf("%s (%s),", stmt, c.ToDistSQL())
	}
	stmt = strings.TrimSuffix(stmt, ",")
	return stmt
}

type Column struct {
	// columnName
	Name Name
	// plainColumnName
	Plain Plain
	// cipherColumnName
	Cipher Cipher
	// assistedQueryColumnName
	AssistedQueryColumn AssistedQueryColumn
	// likeQueryColumnName
	LikeQueryColumn LikeQueryColumn
	// encryptAlgorithmDefinition
	EncryptionAlgorithm *EncryptionAlgorithmType
	// assistedQueryAlgorithmDefinition
	AssistedQueryAlgorithm *AssistedQueryAlgorithmType
	// likeQueryAlgorithmDefinition
	LikeQueryAlgorithm *LikeQueryAlgorithmType
}

func (t *Column) ToDistSQL() string {
	var stmt string
	stmt = t.Name.ToDistSQL()

	if len(t.Plain) != 0 {
		stmt = fmt.Sprintf("%s, %s", stmt, t.Plain.ToDistSQL())
	}
	if len(t.Cipher) != 0 {
		stmt = fmt.Sprintf("%s, %s", stmt, t.Cipher.ToDistSQL())
	}
	if len(t.AssistedQueryColumn) != 0 {
		stmt = fmt.Sprintf("%s, %s", stmt, t.AssistedQueryColumn.ToDistSQL())
	}
	if len(t.LikeQueryColumn) != 0 {
		stmt = fmt.Sprintf("%s, %s", stmt, t.LikeQueryColumn.ToDistSQL())
	}
	if t.EncryptionAlgorithm != nil {
		stmt = fmt.Sprintf("%s, %s", stmt, t.EncryptionAlgorithm.ToDistSQL())
	}
	if t.AssistedQueryAlgorithm != nil {
		stmt = fmt.Sprintf("%s, %s", stmt, t.AssistedQueryAlgorithm.ToDistSQL())
	}
	if t.LikeQueryAlgorithm != nil {
		stmt = fmt.Sprintf("%s, %s", stmt, t.LikeQueryAlgorithm.ToDistSQL())
	}
	stmt = strings.TrimSuffix(stmt, ",")

	return stmt
}

type Name string

func (t *Name) ToDistSQL() string {
	return fmt.Sprintf("NAME=%s", *t)
}

type Plain string

func (t *Plain) ToDistSQL() string {
	return fmt.Sprintf("PLAIN=%s", *t)
}

type Cipher string

func (t *Cipher) ToDistSQL() string {
	return fmt.Sprintf("CIPHER=%s", *t)
}

type AssistedQueryColumn string

func (t *AssistedQueryColumn) ToDistSQL() string {
	return fmt.Sprintf("ASSISTED_QUERY_COLUMN=%s", *t)
}

type LikeQueryColumn string

func (t *LikeQueryColumn) ToDistSQL() string {
	return fmt.Sprintf("LIKE_QUERY_COLUMN=%s", *t)
}

type EncryptionAlgorithmType struct {
	AlgorithmType
}

func (r *EncryptionAlgorithmType) ToDistSQL() string {
	return fmt.Sprintf("ENCRYPT_ALGORITHM(%s)", r.AlgorithmType.ToDistSQL())
}

type AssistedQueryAlgorithmType struct {
	AlgorithmType
}

func (r *AssistedQueryAlgorithmType) ToDistSQL() string {
	return fmt.Sprintf("ASSISTED_QUERY_ALGORITHM(%s)", r.AlgorithmType.ToDistSQL())
}

type LikeQueryAlgorithmType struct {
	AlgorithmType
}

func (r *LikeQueryAlgorithmType) ToDistSQL() string {
	return fmt.Sprintf("LIKE_QUERY_ALGORITHM(%s)", r.AlgorithmType.ToDistSQL())
}

type AlgorithmType struct {
	Name       string
	Properties Properties
}

func (t *AlgorithmType) ToDistSQL() string {
	var stmt string
	if len(t.Properties) == 0 {
		stmt = fmt.Sprintf("NAME='%s'", t.Name)
	} else {
		stmt = fmt.Sprintf("NAME=%s, %s", t.Name, t.Properties.ToDistSQL())
	}

	return fmt.Sprintf("TYPE(%s)", stmt)
}

type Properties map[string]string

func (t *Properties) ToDistSQL() string {
	var stmt string
	for k, v := range *t {
		if len(stmt) != 0 {
			stmt = fmt.Sprintf("%s, '%s'='%s',", stmt, k, v)
		} else {
			stmt = fmt.Sprintf("'%s'='%s',", k, v)
		}
	}
	stmt = strings.TrimSuffix(stmt, ",")
	return fmt.Sprintf("PROPERTIES(%s)", stmt)
}

type QueryWithCipherColumn bool

func (t *QueryWithCipherColumn) ToDistSQL() string {
	if *t {
		return "QUERY_ WITH_CIPHER_COLUMN=true"
	} else {
		return "QUERY_ WITH_CIPHER_COLUMN=false"
	}
}
