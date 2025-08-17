package transaction

import (
	"fmt"
	"log"
)

// Operation represents a reversible operation
type Operation interface {
	Execute() error
	Rollback() error
	Description() string
}

// TransactionManager manages a series of operations that can be rolled back
type TransactionManager struct {
	operations []Operation
	executed   []Operation
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager() *TransactionManager {
	return &TransactionManager{
		operations: make([]Operation, 0),
		executed:   make([]Operation, 0),
	}
}

// AddOperation adds an operation to the transaction
func (tm *TransactionManager) AddOperation(op Operation) {
	tm.operations = append(tm.operations, op)
}

// Execute runs all operations and handles rollback on failure
func (tm *TransactionManager) Execute() error {
	for i, op := range tm.operations {
		if err := op.Execute(); err != nil {
			// Rollback all previously executed operations
			rollbackErr := tm.rollbackExecuted()
			if rollbackErr != nil {
				return fmt.Errorf("operation %d failed: %w, rollback also failed: %v", i, err, rollbackErr)
			}
			return fmt.Errorf("operation %d failed: %w (rollback successful)", i, err)
		}
		tm.executed = append(tm.executed, op)
	}
	return nil
}

// rollbackExecuted rolls back all executed operations in reverse order
func (tm *TransactionManager) rollbackExecuted() error {
	var rollbackErrors []error
	
	// Rollback in reverse order
	for i := len(tm.executed) - 1; i >= 0; i-- {
		op := tm.executed[i]
		if err := op.Rollback(); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("failed to rollback %s: %w", op.Description(), err))
		}
	}
	
	if len(rollbackErrors) > 0 {
		return fmt.Errorf("rollback errors: %v", rollbackErrors)
	}
	
	return nil
}

// FileOperation represents file system operations
type FileOperation struct {
	executeFunc    func() error
	rollbackFunc   func() error
	description    string
}

// NewFileOperation creates a new file operation
func NewFileOperation(description string, execute, rollback func() error) *FileOperation {
	return &FileOperation{
		executeFunc:  execute,
		rollbackFunc: rollback,
		description:  description,
	}
}

func (fo *FileOperation) Execute() error {
	return fo.executeFunc()
}

func (fo *FileOperation) Rollback() error {
	if fo.rollbackFunc != nil {
		return fo.rollbackFunc()
	}
	return nil
}

func (fo *FileOperation) Description() string {
	return fo.description
}

// DatabaseOperation represents database operations
type DatabaseOperation struct {
	executeFunc    func() error
	rollbackFunc   func() error
	description    string
}

// NewDatabaseOperation creates a new database operation
func NewDatabaseOperation(description string, execute, rollback func() error) *DatabaseOperation {
	return &DatabaseOperation{
		executeFunc:  execute,
		rollbackFunc: rollback,
		description:  description,
	}
}

func (do *DatabaseOperation) Execute() error {
	return do.executeFunc()
}

func (do *DatabaseOperation) Rollback() error {
	if do.rollbackFunc != nil {
		return do.rollbackFunc()
	}
	return nil
}

func (do *DatabaseOperation) Description() string {
	return do.description
}

// CompensatingAction represents an action that compensates for another action
type CompensatingAction struct {
	originalAction string
	compensateFunc func() error
	description    string
}

// NewCompensatingAction creates a compensating action
func NewCompensatingAction(originalAction, description string, compensate func() error) *CompensatingAction {
	return &CompensatingAction{
		originalAction: originalAction,
		compensateFunc: compensate,
		description:    description,
	}
}

func (ca *CompensatingAction) Execute() error {
	// Compensating actions are executed during rollback
	return nil
}

func (ca *CompensatingAction) Rollback() error {
	if ca.compensateFunc != nil {
		return ca.compensateFunc()
	}
	return nil
}

func (ca *CompensatingAction) Description() string {
	return fmt.Sprintf("Compensate for %s: %s", ca.originalAction, ca.description)
}

// RecoveryManager handles failed operations and attempts recovery
type RecoveryManager struct {
	failedOperations []FailedOperation
}

// FailedOperation represents an operation that failed and needs recovery
type FailedOperation struct {
	Operation   Operation
	Error       error
	Timestamp   int64
	Attempts    int
	Recoverable bool
}

// NewRecoveryManager creates a new recovery manager
func NewRecoveryManager() *RecoveryManager {
	return &RecoveryManager{
		failedOperations: make([]FailedOperation, 0),
	}
}

// RecordFailure records a failed operation for later recovery
func (rm *RecoveryManager) RecordFailure(op Operation, err error, recoverable bool) {
	failure := FailedOperation{
		Operation:   op,
		Error:       err,
		Timestamp:   getCurrentTimestamp(),
		Attempts:    0,
		Recoverable: recoverable,
	}
	rm.failedOperations = append(rm.failedOperations, failure)
	
	// Log the failure for audit purposes
	log.Printf("Operation failed: %s - Error: %v - Recoverable: %t", op.Description(), err, recoverable)
}

// AttemptRecovery attempts to recover failed operations
func (rm *RecoveryManager) AttemptRecovery() error {
	var recoveryErrors []error
	
	for i := range rm.failedOperations {
		failure := &rm.failedOperations[i]
		
		if !failure.Recoverable || failure.Attempts >= 3 {
			continue
		}
		
		failure.Attempts++
		
		if err := failure.Operation.Execute(); err != nil {
			recoveryErrors = append(recoveryErrors, fmt.Errorf("recovery attempt %d failed for %s: %w", failure.Attempts, failure.Operation.Description(), err))
		} else {
			// Recovery successful, remove from failed operations
			rm.failedOperations = append(rm.failedOperations[:i], rm.failedOperations[i+1:]...)
			log.Printf("Recovery successful for: %s", failure.Operation.Description())
		}
	}
	
	if len(recoveryErrors) > 0 {
		return fmt.Errorf("recovery errors: %v", recoveryErrors)
	}
	
	return nil
}

// GetFailedOperations returns a list of failed operations
func (rm *RecoveryManager) GetFailedOperations() []FailedOperation {
	return rm.failedOperations
}

// CleanupOldFailures removes old failed operations that are no longer recoverable
func (rm *RecoveryManager) CleanupOldFailures(maxAge int64) {
	currentTime := getCurrentTimestamp()
	var validFailures []FailedOperation
	
	for _, failure := range rm.failedOperations {
		if currentTime-failure.Timestamp < maxAge {
			validFailures = append(validFailures, failure)
		}
	}
	
	rm.failedOperations = validFailures
}

// getCurrentTimestamp returns current Unix timestamp
func getCurrentTimestamp() int64 {
	return int64(1000) // Simplified for this example
}