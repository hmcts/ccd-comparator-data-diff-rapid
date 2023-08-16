package comparator

import (
	"ccd-comparator-data-diff-rapid/config"
	"testing"
)

var appConfigs *config.Configurations

func init() {
	appConfigs = config.GetConfigurations("../", "config_test")
}

func TestRuleFactory_GetEnabledRuleList(t *testing.T) {
	appConfigs.Active = "samevalueafterchange,fieldchangecount,unknownrule"
	factory := NewRuleFactory(appConfigs)
	enabledRuleList := factory.GetEnabledRuleList()

	if len(enabledRuleList) != 2 {
		t.Errorf("Expected 2 enabled rules, but got %d", len(enabledRuleList))
	}

	// Check if the enabled rules are the correct types
	_, isSameValueAfterChangeRule := enabledRuleList[0].(*SameValueAfterChangeRule)
	_, isFieldChangeCountRule := enabledRuleList[1].(*FieldChangeCountRule)

	if !isSameValueAfterChangeRule || !isFieldChangeCountRule {
		t.Error("Expected enabled rules to be SameValueAfterChangeRule and FieldChangeCountRule")
	}
}

func TestRuleFactory_GetSingleRule(t *testing.T) {
	appConfigs.Active = "samevalueafterchange"
	factory := NewRuleFactory(appConfigs)
	enabledRuleList := factory.GetEnabledRuleList()

	if len(enabledRuleList) != 1 {
		t.Errorf("Expected 1 enabled rules, but got %d", len(enabledRuleList))
	}

	_, isSameValueAfterChangeRule := enabledRuleList[0].(*SameValueAfterChangeRule)

	if !isSameValueAfterChangeRule {
		t.Error("Expected enabled rules to be SameValueAfterChangeRule")
	}

	_, isFieldChangeCountRule := enabledRuleList[0].(*FieldChangeCountRule)
	if isFieldChangeCountRule {
		t.Error("Expected disabled rules to be FieldChangeCountRule")
	}
}

func TestRuleFactory_CreateEnabledRuleList_UnknownRuleType(t *testing.T) {
	appConfigs.Active = "samevalueafterchange, fieldchangecount, unknownruletype"
	factory := NewRuleFactory(appConfigs)
	enabledRuleList := factory.GetEnabledRuleList()

	if len(enabledRuleList) != 2 {
		t.Errorf("Expected 1 enabled rules, but got %d", len(enabledRuleList))
	}
}

func TestRuleFactory_CreateEnabledRuleList_NoActiveRuleDefined(t *testing.T) {
	activeAnalyzeRules := ""

	factory := &RuleFactory{}
	_, err := factory.createEnabledRuleList(activeAnalyzeRules)

	if err == nil {
		t.Error("Expected an error, but got nil")
	}

	expectedErrorMsg := "No active rule defined"
	if err.Error() != expectedErrorMsg {
		t.Errorf("Expected error message: %s, but got: %s", expectedErrorMsg, err.Error())
	}
}
