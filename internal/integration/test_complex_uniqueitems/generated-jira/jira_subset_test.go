package api

import (
	"testing"

	"github.com/ogen-go/ogen/validate"
)

// T057: JIRA API subset - real-world validation
// This demonstrates that JIRA API v3 operations that were previously
// blocked by complex uniqueItems now work correctly.

func TestWorkflowTransitionRules_JIRAPattern(t *testing.T) {
	// Test pattern from JIRA: workflow transition rules with uniqueItems
	rule1 := WorkflowTransitionRule{
		RuleKey: "com.atlassian.jira.plugin.system.workflow:validators",
		Configuration: NewOptWorkflowTransitionRuleConfiguration(map[string]string{
			"permissionKey": "BROWSE_PROJECTS",
		}),
		ID: NewOptString("rule-1"),
	}

	rule2 := WorkflowTransitionRule{
		RuleKey: "com.atlassian.jira.plugin.system.workflow:validators",
		Configuration: NewOptWorkflowTransitionRuleConfiguration(map[string]string{
			"permissionKey": "CREATE_ISSUES",
		}),
		ID: NewOptString("rule-2"),
	}

	rule3 := WorkflowTransitionRule{
		RuleKey: "com.atlassian.jira.plugin.system.workflow:post-functions",
		Configuration: NewOptWorkflowTransitionRuleConfiguration(map[string]string{
			"event": "issue_created",
		}),
		ID: NewOptString("rule-3"),
	}

	// All rules are different - should pass validation
	validators := []WorkflowTransitionRule{rule1, rule2, rule3}

	err := validateUniqueWorkflowTransitionRule(validators)
	if err != nil {
		t.Errorf("Expected no error for unique JIRA rules, got: %v", err)
	}

	// Add duplicate - should detect it
	duplicateValidators := []WorkflowTransitionRule{rule1, rule2, rule1}

	err = validateUniqueWorkflowTransitionRule(duplicateValidators)
	if err == nil {
		t.Fatal("Expected DuplicateItemsError for duplicate JIRA rules")
	}

	dupErr, ok := err.(*validate.DuplicateItemsError)
	if !ok {
		t.Fatalf("Expected *validate.DuplicateItemsError, got %T", err)
	}

	if dupErr.Indices[0] != 0 || dupErr.Indices[1] != 2 {
		t.Errorf("Expected indices [0, 2], got %v", dupErr.Indices)
	}
}

func TestIssueTypesWorkflowMapping_JIRAPattern(t *testing.T) {
	// Test pattern from JIRA: issue type to workflow mappings with uniqueItems
	mapping1 := IssueTypesWorkflowMapping{
		WorkflowId:          "workflow-software-simplified",
		IssueTypes:          []string{"10001", "10002"}, // Bug, Task
		UpdateDraftIfNeeded: NewOptBool(true),
	}

	mapping2 := IssueTypesWorkflowMapping{
		WorkflowId:          "workflow-classic",
		IssueTypes:          []string{"10003"}, // Story
		UpdateDraftIfNeeded: NewOptBool(false),
	}

	mapping3 := IssueTypesWorkflowMapping{
		WorkflowId:          "workflow-software-simplified",
		IssueTypes:          []string{"10004", "10005"}, // Epic, Subtask
		UpdateDraftIfNeeded: NewOptBool(true),
	}

	// All mappings are different - should pass
	mappings := []IssueTypesWorkflowMapping{mapping1, mapping2, mapping3}

	err := validateUniqueIssueTypesWorkflowMapping(mappings)
	if err != nil {
		t.Errorf("Expected no error for unique JIRA mappings, got: %v", err)
	}

	// Duplicate mapping - should detect
	duplicateMappings := []IssueTypesWorkflowMapping{mapping1, mapping2, mapping1}

	err = validateUniqueIssueTypesWorkflowMapping(duplicateMappings)
	if err == nil {
		t.Fatal("Expected DuplicateItemsError for duplicate JIRA mappings")
	}

	dupErr, ok := err.(*validate.DuplicateItemsError)
	if !ok {
		t.Fatalf("Expected *validate.DuplicateItemsError, got %T", err)
	}

	if dupErr.Indices[0] != 0 || dupErr.Indices[1] != 2 {
		t.Errorf("Expected indices [0, 2], got %v", dupErr.Indices)
	}
}

func TestWorkflowTransitionRulesUpdate_CompleteJIRARequest(t *testing.T) {
	// Simulate a complete JIRA API request with multiple uniqueItems arrays
	postFunction1 := WorkflowTransitionRule{
		RuleKey: "com.atlassian.jira.plugin.system.workflow:post-functions",
		ID:      NewOptString("pf-1"),
	}

	postFunction2 := WorkflowTransitionRule{
		RuleKey: "com.atlassian.jira.plugin.system.workflow:fire-event",
		ID:      NewOptString("pf-2"),
	}

	condition1 := WorkflowTransitionRule{
		RuleKey: "com.atlassian.jira.plugin.system.workflow:conditions",
		ID:      NewOptString("cond-1"),
	}

	validator1 := WorkflowTransitionRule{
		RuleKey: "com.atlassian.jira.plugin.system.workflow:validators",
		ID:      NewOptString("val-1"),
	}

	rules := WorkflowTransitionRules{
		WorkflowId:    "workflow-1",
		PostFunctions: []WorkflowTransitionRule{postFunction1, postFunction2},
		Conditions:    []WorkflowTransitionRule{condition1},
		Validators:    []WorkflowTransitionRule{validator1},
	}

	// Validate the complete structure
	err := rules.Validate()
	if err != nil {
		t.Errorf("Expected no error for valid JIRA request, got: %v", err)
	}

	// Test with duplicate in postFunctions
	rulesWithDup := WorkflowTransitionRules{
		WorkflowId:    "workflow-1",
		PostFunctions: []WorkflowTransitionRule{postFunction1, postFunction1}, // Duplicate
		Conditions:    []WorkflowTransitionRule{condition1},
		Validators:    []WorkflowTransitionRule{validator1},
	}

	err = rulesWithDup.Validate()
	if err == nil {
		t.Fatal("Expected validation error for duplicate postFunctions")
	}

	t.Logf("Correctly detected duplicate: %v", err)
}

func TestWorkflowSchemeAssociations_CompleteJIRARequest(t *testing.T) {
	// Test the workflow scheme mapping endpoint
	mapping1 := IssueTypesWorkflowMapping{
		WorkflowId: "workflow-simple",
		IssueTypes: []string{"10001", "10002"},
	}

	mapping2 := IssueTypesWorkflowMapping{
		WorkflowId: "workflow-advanced",
		IssueTypes: []string{"10003"},
	}

	associations := WorkflowSchemeAssociations{
		IssueTypeMappings: []IssueTypesWorkflowMapping{mapping1, mapping2},
		DefaultWorkflowId: NewOptString("workflow-default"),
	}

	// Validate complete structure
	err := associations.Validate()
	if err != nil {
		t.Errorf("Expected no error for valid JIRA associations, got: %v", err)
	}

	// Test with duplicate mappings
	associationsWithDup := WorkflowSchemeAssociations{
		IssueTypeMappings: []IssueTypesWorkflowMapping{mapping1, mapping1}, // Duplicate
		DefaultWorkflowId: NewOptString("workflow-default"),
	}

	err = associationsWithDup.Validate()
	if err == nil {
		t.Fatal("Expected validation error for duplicate mappings")
	}

	t.Logf("Correctly detected duplicate: %v", err)
}

// Benchmark JIRA-like workload
func BenchmarkJIRAWorkflowRules_50Rules(b *testing.B) {
	rules := make([]WorkflowTransitionRule, 50)
	for i := 0; i < 50; i++ {
		rules[i] = WorkflowTransitionRule{
			RuleKey: string(rune('A' + (i % 26))),
			ID:      NewOptString(string(rune('0' + (i % 10)))),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := validateUniqueWorkflowTransitionRule(rules)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}
