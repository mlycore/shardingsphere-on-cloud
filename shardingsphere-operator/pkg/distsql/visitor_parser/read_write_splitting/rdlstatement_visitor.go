// Code generated from RDLStatement.g4 by ANTLR 4.8. DO NOT EDIT.

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

package parser // RDLStatement

import "github.com/antlr/antlr4/runtime/Go/antlr"

// A complete Visitor for a parse tree produced by RDLStatementParser.
type RDLStatementVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by RDLStatementParser#createReadwriteSplittingRule.
	VisitCreateReadwriteSplittingRule(ctx *CreateReadwriteSplittingRuleContext) interface{}

	// Visit a parse tree produced by RDLStatementParser#alterReadwriteSplittingRule.
	VisitAlterReadwriteSplittingRule(ctx *AlterReadwriteSplittingRuleContext) interface{}

	// Visit a parse tree produced by RDLStatementParser#dropReadwriteSplittingRule.
	VisitDropReadwriteSplittingRule(ctx *DropReadwriteSplittingRuleContext) interface{}

	// Visit a parse tree produced by RDLStatementParser#readwriteSplittingRuleDefinition.
	VisitReadwriteSplittingRuleDefinition(ctx *ReadwriteSplittingRuleDefinitionContext) interface{}

	// Visit a parse tree produced by RDLStatementParser#dataSourceDefinition.
	VisitDataSourceDefinition(ctx *DataSourceDefinitionContext) interface{}

	// Visit a parse tree produced by RDLStatementParser#ruleName.
	VisitRuleName(ctx *RuleNameContext) interface{}

	// Visit a parse tree produced by RDLStatementParser#writeStorageUnit.
	VisitWriteStorageUnit(ctx *WriteStorageUnitContext) interface{}

	// Visit a parse tree produced by RDLStatementParser#readStorageUnits.
	VisitReadStorageUnits(ctx *ReadStorageUnitsContext) interface{}

	// Visit a parse tree produced by RDLStatementParser#transactionalReadQueryStrategy.
	VisitTransactionalReadQueryStrategy(ctx *TransactionalReadQueryStrategyContext) interface{}

	// Visit a parse tree produced by RDLStatementParser#writeStorageUnitName.
	VisitWriteStorageUnitName(ctx *WriteStorageUnitNameContext) interface{}

	// Visit a parse tree produced by RDLStatementParser#readStorageUnitsNames.
	VisitReadStorageUnitsNames(ctx *ReadStorageUnitsNamesContext) interface{}

	// Visit a parse tree produced by RDLStatementParser#transactionalReadQueryStrategyName.
	VisitTransactionalReadQueryStrategyName(ctx *TransactionalReadQueryStrategyNameContext) interface{}

	// Visit a parse tree produced by RDLStatementParser#ifExists.
	VisitIfExists(ctx *IfExistsContext) interface{}

	// Visit a parse tree produced by RDLStatementParser#ifNotExists.
	VisitIfNotExists(ctx *IfNotExistsContext) interface{}

	// Visit a parse tree produced by RDLStatementParser#literal.
	VisitLiteral(ctx *LiteralContext) interface{}

	// Visit a parse tree produced by RDLStatementParser#algorithmDefinition.
	VisitAlgorithmDefinition(ctx *AlgorithmDefinitionContext) interface{}

	// Visit a parse tree produced by RDLStatementParser#algorithmTypeName.
	VisitAlgorithmTypeName(ctx *AlgorithmTypeNameContext) interface{}

	// Visit a parse tree produced by RDLStatementParser#buildInReadQueryLoadBalanceAlgorithmType.
	VisitBuildInReadQueryLoadBalanceAlgorithmType(ctx *BuildInReadQueryLoadBalanceAlgorithmTypeContext) interface{}

	// Visit a parse tree produced by RDLStatementParser#propertiesDefinition.
	VisitPropertiesDefinition(ctx *PropertiesDefinitionContext) interface{}

	// Visit a parse tree produced by RDLStatementParser#properties.
	VisitProperties(ctx *PropertiesContext) interface{}

	// Visit a parse tree produced by RDLStatementParser#property.
	VisitProperty(ctx *PropertyContext) interface{}

	// Visit a parse tree produced by RDLStatementParser#databaseName.
	VisitDatabaseName(ctx *DatabaseNameContext) interface{}

	// Visit a parse tree produced by RDLStatementParser#groupName.
	VisitGroupName(ctx *GroupNameContext) interface{}

	// Visit a parse tree produced by RDLStatementParser#storageUnitName.
	VisitStorageUnitName(ctx *StorageUnitNameContext) interface{}
}
