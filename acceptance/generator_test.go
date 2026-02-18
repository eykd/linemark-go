package acceptance

import (
	"strings"
	"testing"
)

func TestGenerateTests_SingleScenario(t *testing.T) {
	feature := &Feature{
		SourceFile: "specs/US1-add-item.txt",
		Scenarios: []Scenario{
			{
				Description: "User can add a new outline item.",
				Steps: []Step{
					{Keyword: "GIVEN", Text: "an empty outline.", Line: 4},
					{Keyword: "WHEN", Text: `the user adds an item titled "Chapter 1".`, Line: 6},
					{Keyword: "THEN", Text: "the outline contains 1 item.", Line: 8},
				},
				Line: 2,
			},
		},
	}

	output, err := GenerateTests(feature)
	if err != nil {
		t.Fatalf("GenerateTests() error = %v", err)
	}

	// Must contain DO NOT EDIT marker
	if !strings.Contains(output, "DO NOT EDIT") {
		t.Error("output missing 'DO NOT EDIT' marker")
	}

	// Must contain package declaration
	if !strings.Contains(output, "package acceptance_test") {
		t.Error("output missing package declaration")
	}

	// Must import testing
	if !strings.Contains(output, `"testing"`) {
		t.Error("output missing testing import")
	}

	// Must contain a test function
	if !strings.Contains(output, "func Test") {
		t.Error("output missing test function")
	}

	// Must contain t.Fatal stub
	if !strings.Contains(output, "t.Fatal") {
		t.Error("output missing t.Fatal stub")
	}

	// Must reference the scenario description
	if !strings.Contains(output, "User can add a new outline item") {
		t.Error("output missing scenario description")
	}

	// Must contain step comments
	if !strings.Contains(output, "GIVEN") {
		t.Error("output missing GIVEN step comment")
	}
	if !strings.Contains(output, "WHEN") {
		t.Error("output missing WHEN step comment")
	}
	if !strings.Contains(output, "THEN") {
		t.Error("output missing THEN step comment")
	}
}

func TestGenerateTests_MultipleScenarios(t *testing.T) {
	feature := &Feature{
		SourceFile: "specs/US2-manage.txt",
		Scenarios: []Scenario{
			{
				Description: "User can add an item.",
				Steps: []Step{
					{Keyword: "GIVEN", Text: "an empty outline.", Line: 4},
				},
				Line: 2,
			},
			{
				Description: "User can remove an item.",
				Steps: []Step{
					{Keyword: "GIVEN", Text: "an outline with items.", Line: 12},
				},
				Line: 10,
			},
		},
	}

	output, err := GenerateTests(feature)
	if err != nil {
		t.Fatalf("GenerateTests() error = %v", err)
	}

	// Must have two test functions
	count := strings.Count(output, "func Test")
	if count != 2 {
		t.Errorf("test function count = %d, want 2", count)
	}

	if !strings.Contains(output, "User can add an item") {
		t.Error("output missing first scenario description")
	}
	if !strings.Contains(output, "User can remove an item") {
		t.Error("output missing second scenario description")
	}
}

func TestGenerateTests_EmptyFeature(t *testing.T) {
	feature := &Feature{
		SourceFile: "specs/empty.txt",
	}

	output, err := GenerateTests(feature)
	if err != nil {
		t.Fatalf("GenerateTests() error = %v", err)
	}

	// Should still produce valid Go with DO NOT EDIT marker
	if !strings.Contains(output, "DO NOT EDIT") {
		t.Error("output missing 'DO NOT EDIT' marker")
	}
	if !strings.Contains(output, "package acceptance_test") {
		t.Error("output missing package declaration")
	}

	// Should not contain test functions
	if strings.Contains(output, "func Test") {
		t.Error("output should not contain test functions for empty feature")
	}
}

func TestGenerateTests_FunctionNameSanitization(t *testing.T) {
	feature := &Feature{
		SourceFile: "specs/US1-test.txt",
		Scenarios: []Scenario{
			{
				Description: `User adds "special" item/thing.`,
				Steps: []Step{
					{Keyword: "GIVEN", Text: "something.", Line: 4},
				},
				Line: 2,
			},
		},
	}

	output, err := GenerateTests(feature)
	if err != nil {
		t.Fatalf("GenerateTests() error = %v", err)
	}

	// Function name should not contain special chars
	// Should contain a valid Go test function name
	if !strings.Contains(output, "func Test") {
		t.Error("output missing test function")
	}

	// Function name should not have quotes, slashes, or dots
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "func Test") {
			// Extract just the function name (before the parenthesis)
			parenIdx := strings.Index(trimmed, "(")
			if parenIdx == -1 {
				continue
			}
			funcName := trimmed[len("func "):parenIdx]
			if strings.ContainsAny(funcName, `"/.`) {
				t.Errorf("function name contains invalid chars: %s", funcName)
			}
		}
	}
}

func TestGenerateTests_SourceFileInComment(t *testing.T) {
	feature := &Feature{
		SourceFile: "specs/US5-feature.txt",
		Scenarios: []Scenario{
			{
				Description: "A scenario.",
				Steps: []Step{
					{Keyword: "GIVEN", Text: "something.", Line: 4},
				},
				Line: 2,
			},
		},
	}

	output, err := GenerateTests(feature)
	if err != nil {
		t.Fatalf("GenerateTests() error = %v", err)
	}

	if !strings.Contains(output, "specs/US5-feature.txt") {
		t.Error("output missing source file reference")
	}
}

func TestGenerateTests_StepTextInComments(t *testing.T) {
	feature := &Feature{
		SourceFile: "specs/US1-test.txt",
		Scenarios: []Scenario{
			{
				Description: "Test scenario.",
				Steps: []Step{
					{Keyword: "GIVEN", Text: "a precondition exists.", Line: 4},
					{Keyword: "WHEN", Text: "the action is performed.", Line: 6},
					{Keyword: "THEN", Text: "the expected outcome occurs.", Line: 8},
				},
				Line: 2,
			},
		},
	}

	output, err := GenerateTests(feature)
	if err != nil {
		t.Fatalf("GenerateTests() error = %v", err)
	}

	if !strings.Contains(output, "a precondition exists.") {
		t.Error("output missing GIVEN step text")
	}
	if !strings.Contains(output, "the action is performed.") {
		t.Error("output missing WHEN step text")
	}
	if !strings.Contains(output, "the expected outcome occurs.") {
		t.Error("output missing THEN step text")
	}
}
