package balogan

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"
)

func TestCondition_Always(t *testing.T) {
	condition := Always()

	for range 10 {
		if !condition() {
			t.Error("Always() should always return true")
		}
	}
}

func TestCondition_Never(t *testing.T) {
	condition := Never()

	for range 10 {
		if condition() {
			t.Error("Never() should always return false")
		}
	}
}

func TestCondition_EnvEquals(t *testing.T) {
	condition := EnvEquals("TEST_BALOGAN_VAR", "test_value")
	if condition() {
		t.Error("EnvEquals should return false for non-existent environment variable")
	}

	os.Setenv("TEST_BALOGAN_VAR", "test_value")
	defer os.Unsetenv("TEST_BALOGAN_VAR")

	if !condition() {
		t.Error("EnvEquals should return true when environment variable matches")
	}

	wrongCondition := EnvEquals("TEST_BALOGAN_VAR", "wrong_value")
	if wrongCondition() {
		t.Error("EnvEquals should return false when environment variable doesn't match")
	}
}

func TestCondition_EnvExists(t *testing.T) {
	condition := EnvExists("TEST_BALOGAN_EXISTS")
	if condition() {
		t.Error("EnvExists should return false for non-existent environment variable")
	}

	os.Setenv("TEST_BALOGAN_EXISTS", "")
	defer os.Unsetenv("TEST_BALOGAN_EXISTS")

	if !condition() {
		t.Error("EnvExists should return true when environment variable exists (even if empty)")
	}

	os.Setenv("TEST_BALOGAN_EXISTS", "some_value")
	if !condition() {
		t.Error("EnvExists should return true when environment variable exists with value")
	}
}

func TestCondition_RandomSample(t *testing.T) {
	neverCondition := RandomSample(0)
	if neverCondition() {
		t.Error("RandomSample(0) should never return true")
	}

	negativeCondition := RandomSample(-10)
	if negativeCondition() {
		t.Error("RandomSample with negative percentage should never return true")
	}

	alwaysCondition := RandomSample(100)
	if !alwaysCondition() {
		t.Error("RandomSample(100) should always return true")
	}

	overCondition := RandomSample(150)
	if !overCondition() {
		t.Error("RandomSample over 100% should always return true")
	}

	sampleCondition := RandomSample(50)

	results := make(map[bool]int)
	totalTests := 100

	for range totalTests {
		time.Sleep(1 * time.Microsecond)
		result := sampleCondition()
		results[result]++
	}

	if results[true] == 0 {
		t.Error("RandomSample(50) should return true at least sometimes")
	}
	if results[false] == 0 {
		t.Error("RandomSample(50) should return false at least sometimes")
	}
	if results[true] == totalTests {
		t.Error("RandomSample(50) should not always return true")
	}
	if results[false] == totalTests {
		t.Error("RandomSample(50) should not always return false")
	}
}

func TestCondition_RateLimit(t *testing.T) {
	neverCondition := RateLimit(0)
	if neverCondition() {
		t.Error("RateLimit(0) should never return true")
	}

	negativeCondition := RateLimit(-5)
	if negativeCondition() {
		t.Error("RateLimit with negative value should never return true")
	}

	limitCondition := RateLimit(3)

	for i := range 3 {
		if !limitCondition() {
			t.Errorf("RateLimit(3) call %d should return true", i+1)
		}
	}

	for i := range 5 {
		if limitCondition() {
			t.Errorf("RateLimit(3) call %d should return false (exceeded limit)", i+4)
		}
	}

	time.Sleep(1100 * time.Millisecond)

	if !limitCondition() {
		t.Error("RateLimit should reset after 1 second")
	}
}

func TestCondition_RateLimitConcurrency(t *testing.T) {
	limitCondition := RateLimit(10)

	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0

	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if limitCondition() {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if successCount != 10 {
		t.Errorf("RateLimit(10) with 50 concurrent calls should allow exactly 10, got %d", successCount)
	}
}

func TestCondition_TimeRange(t *testing.T) {
	workHoursCondition := TimeRange(9, 17)

	result1 := workHoursCondition()
	result2 := workHoursCondition()

	if result1 != result2 {
		t.Error("TimeRange should return consistent results when called immediately")
	}

	nightCondition := TimeRange(22, 6)
	nightResult1 := nightCondition()
	nightResult2 := nightCondition()

	if nightResult1 != nightResult2 {
		t.Error("TimeRange with overnight range should return consistent results")
	}
}

func TestCondition_HasContextValue(t *testing.T) {
	type testKey string
	const myTestKey testKey = "test_key"

	condition := HasContextValue(myTestKey)

	if condition(nil) {
		t.Error("HasContextValue should return false for nil context")
	}

	ctx := context.Background()
	if condition(ctx) {
		t.Error("HasContextValue should return false when key doesn't exist")
	}

	ctxWithValue := context.WithValue(ctx, myTestKey, "test_value")
	if !condition(ctxWithValue) {
		t.Error("HasContextValue should return true when key exists")
	}

	ctxWithNil := context.WithValue(ctx, myTestKey, nil)
	if condition(ctxWithNil) {
		t.Error("HasContextValue should return false when key exists but value is nil")
	}
}

func TestCondition_ContextValueEquals(t *testing.T) {
	type testKey string
	const myTestKey testKey = "test_key"

	condition := ContextValueEquals(myTestKey, "expected_value")

	if condition(nil) {
		t.Error("ContextValueEquals should return false for nil context")
	}

	ctx := context.Background()
	if condition(ctx) {
		t.Error("ContextValueEquals should return false when key doesn't exist")
	}

	ctxWithCorrectValue := context.WithValue(ctx, myTestKey, "expected_value")
	if !condition(ctxWithCorrectValue) {
		t.Error("ContextValueEquals should return true when value matches")
	}

	ctxWithWrongValue := context.WithValue(ctx, myTestKey, "wrong_value")
	if condition(ctxWithWrongValue) {
		t.Error("ContextValueEquals should return false when value doesn't match")
	}
}

func TestCondition_OnlyLevel(t *testing.T) {
	condition := OnlyLevel(ErrorLevel)
	fields := Fields{"test": "value"}

	if !condition(ErrorLevel, fields) {
		t.Error("OnlyLevel should return true for matching level")
	}

	if condition(InfoLevel, fields) {
		t.Error("OnlyLevel should return false for non-matching level")
	}

	if condition(WarningLevel, fields) {
		t.Error("OnlyLevel should return false for non-matching level")
	}
}

func TestCondition_MinLevel(t *testing.T) {
	condition := MinLevel(WarningLevel)
	fields := Fields{"test": "value"}

	if !condition(WarningLevel, fields) {
		t.Error("MinLevel should return true for exact minimum level")
	}

	if !condition(ErrorLevel, fields) {
		t.Error("MinLevel should return true for level above minimum")
	}

	if !condition(FatalLevel, fields) {
		t.Error("MinLevel should return true for level above minimum")
	}

	if condition(InfoLevel, fields) {
		t.Error("MinLevel should return false for level below minimum")
	}

	if condition(DebugLevel, fields) {
		t.Error("MinLevel should return false for level below minimum")
	}
}

func TestCondition_HasField(t *testing.T) {
	condition := HasField("user_id")

	fieldsWithUser := Fields{"user_id": 123, "other": "value"}
	if !condition(InfoLevel, fieldsWithUser) {
		t.Error("HasField should return true when field exists")
	}

	fieldsWithoutUser := Fields{"other": "value"}
	if condition(InfoLevel, fieldsWithoutUser) {
		t.Error("HasField should return false when field doesn't exist")
	}

	emptyFields := Fields{}
	if condition(InfoLevel, emptyFields) {
		t.Error("HasField should return false for empty fields")
	}
}

func TestCondition_FieldEquals(t *testing.T) {
	condition := FieldEquals("environment", "production")

	matchingFields := Fields{"environment": "production", "other": "value"}
	if !condition(InfoLevel, matchingFields) {
		t.Error("FieldEquals should return true when field value matches")
	}

	nonMatchingFields := Fields{"environment": "development", "other": "value"}
	if condition(InfoLevel, nonMatchingFields) {
		t.Error("FieldEquals should return false when field value doesn't match")
	}

	missingFields := Fields{"other": "value"}
	if condition(InfoLevel, missingFields) {
		t.Error("FieldEquals should return false when field doesn't exist")
	}
}

func TestCondition_And(t *testing.T) {
	trueCondition := func() bool { return true }
	falseCondition := func() bool { return false }

	allTrue := And(trueCondition, trueCondition, trueCondition)
	if !allTrue() {
		t.Error("And should return true when all conditions are true")
	}

	oneFalse := And(trueCondition, falseCondition, trueCondition)
	if oneFalse() {
		t.Error("And should return false when any condition is false")
	}

	allFalse := And(falseCondition, falseCondition, falseCondition)
	if allFalse() {
		t.Error("And should return false when all conditions are false")
	}

	noConditions := And()
	if !noConditions() {
		t.Error("And with no conditions should return true")
	}
}

func TestCondition_Or(t *testing.T) {
	trueCondition := func() bool { return true }
	falseCondition := func() bool { return false }

	allTrue := Or(trueCondition, trueCondition, trueCondition)
	if !allTrue() {
		t.Error("Or should return true when all conditions are true")
	}

	oneTrue := Or(falseCondition, trueCondition, falseCondition)
	if !oneTrue() {
		t.Error("Or should return true when any condition is true")
	}

	allFalse := Or(falseCondition, falseCondition, falseCondition)
	if allFalse() {
		t.Error("Or should return false when all conditions are false")
	}

	noConditions := Or()
	if noConditions() {
		t.Error("Or with no conditions should return false")
	}
}

func TestCondition_Not(t *testing.T) {
	trueCondition := func() bool { return true }
	falseCondition := func() bool { return false }

	notTrue := Not(trueCondition)
	if notTrue() {
		t.Error("Not should return false when negating true condition")
	}

	notFalse := Not(falseCondition)
	if !notFalse() {
		t.Error("Not should return true when negating false condition")
	}
}

func TestCondition_Any(t *testing.T) {
	trueCondition := func() bool { return true }
	falseCondition := func() bool { return false }

	anyTrue := Any(falseCondition, trueCondition, falseCondition)
	if !anyTrue() {
		t.Error("Any should return true when any condition is true")
	}

	anyFalse := Any(falseCondition, falseCondition, falseCondition)
	if anyFalse() {
		t.Error("Any should return false when all conditions are false")
	}
}

func TestCondition_All(t *testing.T) {
	trueCondition := func() bool { return true }
	falseCondition := func() bool { return false }

	allTrue := All(trueCondition, trueCondition, trueCondition)
	if !allTrue() {
		t.Error("All should return true when all conditions are true")
	}

	allFalse := All(trueCondition, falseCondition, trueCondition)
	if allFalse() {
		t.Error("All should return false when any condition is false")
	}
}

func TestCondition_CountBased(t *testing.T) {
	condition := CountBased(3)

	for i := range 3 {
		if !condition() {
			t.Errorf("CountBased(3) call %d should return true", i+1)
		}
	}

	for i := range 5 {
		if condition() {
			t.Errorf("CountBased(3) call %d should return false (exceeded count)", i+4)
		}
	}
}

func TestCondition_CountBasedConcurrency(t *testing.T) {
	condition := CountBased(10)

	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0

	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if condition() {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if successCount != 10 {
		t.Errorf("CountBased(10) with 50 concurrent calls should allow exactly 10, got %d", successCount)
	}
}

func TestCondition_SampleEveryN(t *testing.T) {
	alwaysCondition := SampleEveryN(1)
	for range 5 {
		if !alwaysCondition() {
			t.Error("SampleEveryN(1) should always return true")
		}
	}

	zeroCondition := SampleEveryN(0)
	for range 5 {
		if !zeroCondition() {
			t.Error("SampleEveryN(0) should always return true (same as SampleEveryN(1))")
		}
	}

	condition := SampleEveryN(3)

	if !condition() {
		t.Error("SampleEveryN(3) first call should return true")
	}

	if condition() {
		t.Error("SampleEveryN(3) second call should return false")
	}

	if condition() {
		t.Error("SampleEveryN(3) third call should return false")
	}

	if !condition() {
		t.Error("SampleEveryN(3) fourth call should return true")
	}
}

func TestCondition_PredefinedEnvironmentConditions(t *testing.T) {
	os.Setenv("ENV", "production")
	defer os.Unsetenv("ENV")

	if !InProduction() {
		t.Error("InProduction should return true when ENV=production")
	}

	if InDevelopment() {
		t.Error("InDevelopment should return false when ENV=production")
	}

	os.Setenv("ENV", "development")

	if !InDevelopment() {
		t.Error("InDevelopment should return true when ENV=development")
	}

	if InProduction() {
		t.Error("InProduction should return false when ENV=development")
	}

	os.Setenv("ENV", "test")

	if !InTesting() {
		t.Error("InTesting should return true when ENV=test")
	}

	os.Setenv("ENV", "staging")

	if !InStaging() {
		t.Error("InStaging should return true when ENV=staging")
	}
}

func TestCondition_PredefinedDebugConditions(t *testing.T) {
	os.Setenv("DEBUG", "true")
	defer os.Unsetenv("DEBUG")

	if !DebugEnabled() {
		t.Error("DebugEnabled should return true when DEBUG=true")
	}

	os.Setenv("DEBUG", "false")
	if DebugEnabled() {
		t.Error("DebugEnabled should return false when DEBUG=false")
	}

	os.Setenv("VERBOSE", "true")
	defer os.Unsetenv("VERBOSE")

	if !VerboseMode() {
		t.Error("VerboseMode should return true when VERBOSE=true")
	}

	os.Setenv("VERBOSE", "false")
	if VerboseMode() {
		t.Error("VerboseMode should return false when VERBOSE=false")
	}
}

func TestCondition_PredefinedTimeConditions(t *testing.T) {
	result1 := WorkingHours()
	result2 := WorkingHours()

	if result1 != result2 {
		t.Error("WorkingHours should return consistent results when called immediately")
	}

	weekend := Weekend()
	weekday := Weekday()

	if weekend == weekday {
		t.Error("Weekend and Weekday should return opposite values")
	}

	if Weekend() != weekend {
		t.Error("Weekend should return consistent results")
	}

	if Weekday() != weekday {
		t.Error("Weekday should return consistent results")
	}
}

func TestCondition_ComplexConditionCombinations(t *testing.T) {
	os.Setenv("ENV", "development")
	os.Setenv("DEBUG", "true")
	defer func() {
		os.Unsetenv("ENV")
		os.Unsetenv("DEBUG")
	}()

	complexCondition := And(
		Or(InDevelopment, InTesting),
		DebugEnabled,
	)

	if !complexCondition() {
		t.Error("Complex condition should return true for development environment with debug enabled")
	}

	os.Setenv("ENV", "production")

	if complexCondition() {
		t.Error("Complex condition should return false for production environment")
	}

	notProductionCondition := Not(InProduction)

	if notProductionCondition() {
		t.Error("Not(InProduction) should return false when ENV=production")
	}
}

func BenchmarkCondition_Always(b *testing.B) {
	condition := Always()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		condition()
	}
}

func BenchmarkCondition_RateLimit(b *testing.B) {
	condition := RateLimit(1000000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		condition()
	}
}

func BenchmarkCondition_RandomSample(b *testing.B) {
	condition := RandomSample(50)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		condition()
	}
}

func BenchmarkCondition_ComplexCondition(b *testing.B) {
	condition := And(
		Or(InDevelopment, InTesting),
		DebugEnabled,
		Not(Weekend),
	)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		condition()
	}
}
