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
	Schemas []Schema `json:"schemas"`
}

type Schema struct {
	Name       string          `json:"name"`
	Encryption EncryptRule     `json:"encryption"`
	Topology   ClusterTopology `json:"topology"`
}

type ClusterTopology struct {
	ComputeNode ComputeNodeSpec `json:"computeNode"`
	StorageNode StorageNodeSpec `json:"storageNode"`
}

// type CreateEncryptRule struct {
// 	EncryptRule
// }

// func (t *CreateEncryptRule) ToDistSQL() string {
// 	return fmt.Sprintf("CREATE ENCRYPT RULE %s", t.EncryptRule.ToDistSQL())
// }

type EncryptRule struct {
	// ifNotExists
	IfNotExists IfNotExists `json:"ifNotExists,omitempty"`
	// IfNotExist bool
	// encryptDefinition
	EncryptDefinitions EncryptDefinitionList `json:"encryptDefinitions"`
}

func (t *EncryptRule) ToDistSQL() string {
	var stmt string
	if t.IfNotExists {
		stmt = fmt.Sprintf("%s", t.IfNotExists.ToDistSQL())
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
	Name string `json:"name"`
	// columns
	Columns ColumnList `json:"columns"`
	// queryWithCipheColumn
	QueryWithCipherColumn QueryWithCipherColumn `json:"queryWithCipherColumn"`
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
	Name Name `json:"name"`
	// plainColumnName
	Plain Plain `json:"plain,omitempty"`
	// cipherColumnName
	Cipher Cipher `json:"cipher,omitempty"`
	// assistedQueryColumnName
	AssistedQueryColumn AssistedQueryColumn `json:"assistedQueryColumn,omitempty"`
	// likeQueryColumnName
	LikeQueryColumn LikeQueryColumn `json:"likeQueryColumn,omitempty"`
	// encryptAlgorithmDefinition
	EncryptionAlgorithm *EncryptionAlgorithmType `json:"encryptionAlgorithm,omitempty"`
	// assistedQueryAlgorithmDefinition
	AssistedQueryAlgorithm *AssistedQueryAlgorithmType `json:"assistedQueryAlgorithm,omitempty"`
	// likeQueryAlgorithmDefinition
	LikeQueryAlgorithm *LikeQueryAlgorithmType `json:"likeQueryAlgorithm,omitempty"`
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

// type EncryptionAlgorithmType struct {
// 	*AlgorithmType
// }

type EncryptionAlgorithmType AlgorithmType

func (r *EncryptionAlgorithmType) ToDistSQL() string {
	return fmt.Sprintf("ENCRYPT_ALGORITHM(%s)", r.ToDistSQL())
}

// type AssistedQueryAlgorithmType struct {
// 	*AlgorithmType
// }

type AssistedQueryAlgorithmType AlgorithmType

func (r *AssistedQueryAlgorithmType) ToDistSQL() string {
	return fmt.Sprintf("ASSISTED_QUERY_ALGORITHM(%s)", r.ToDistSQL())
}

// type LikeQueryAlgorithmType struct {
// 	*AlgorithmType
// }

type LikeQueryAlgorithmType AlgorithmType

func (r *LikeQueryAlgorithmType) ToDistSQL() string {
	return fmt.Sprintf("LIKE_QUERY_ALGORITHM(%s)", r.ToDistSQL())
}

type AlgorithmType struct {
	Name       string     `json:"name"`
	Properties Properties `json:"properties,omitempty"`
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

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
