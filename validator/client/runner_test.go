package client

import (
	"context"
	"errors"
	"testing"

	pb "github.com/prysmaticlabs/prysm/proto/beacon/rpc/v1"
	"github.com/prysmaticlabs/prysm/shared/testutil"
	logTest "github.com/sirupsen/logrus/hooks/test"
)

func cancelledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func TestRunInitializesValidator(t *testing.T) {
	v := &fakeValidator{}
	run(cancelledContext(), v)
	if !v.InitializeCalled {
		t.Error("Expected Initialize() to be called")
	}
}

func TestRunCleansUpValidator(t *testing.T) {
	v := &fakeValidator{}
	run(cancelledContext(), v)
	if !v.DoneCalled {
		t.Error("Expected Done() to be called")
	}
}

func TestRunWaitsForActivation(t *testing.T) {
	v := &fakeValidator{}
	run(cancelledContext(), v)
	if !v.WaitForActivationCalled {
		t.Error("Expected WaitForActivation() to be called")
	}
}

func TestRunOnNextSlotUpdatesAssignments(t *testing.T) {
	v := &fakeValidator{}
	ctx, cancel := context.WithCancel(context.Background())

	slot := uint64(55)
	ticker := make(chan uint64)
	v.NextSlotRet = ticker
	go func() {
		ticker <- slot

		cancel()
	}()

	run(ctx, v)

	if !v.UpdateAssignmentsCalled {
		t.Fatalf("Expected UpdateAssignments(%d) to be called", slot)
	}
	if v.UpdateAssignmentsArg1 != slot {
		t.Errorf("UpdateAssignments was called with wrong argument. Want=%d, got=%d", slot, v.UpdateAssignmentsArg1)
	}
}

func TestRunOnNextSlotUpdatesAssignmentsHandlesError(t *testing.T) {
	hook := logTest.NewGlobal()
	v := &fakeValidator{}
	ctx, cancel := context.WithCancel(context.Background())

	slot := uint64(55)
	ticker := make(chan uint64)
	v.NextSlotRet = ticker
	go func() {
		ticker <- slot

		cancel()
	}()
	v.UpdateAssignmentsRet = errors.New("bad")

	run(ctx, v)

	testutil.AssertLogsContain(t, hook, "Failed to update assignments")
}

func TestRunOnNextSlotDeterminesRole(t *testing.T) {
	v := &fakeValidator{}
	ctx, cancel := context.WithCancel(context.Background())

	slot := uint64(55)
	ticker := make(chan uint64)
	v.NextSlotRet = ticker
	go func() {
		ticker <- slot

		cancel()
	}()

	run(ctx, v)

	if !v.RoleAtCalled {
		t.Fatalf("Expected RoleAt(%d) to be called", slot)
	}
	if v.RoleAtArg1 != slot {
		t.Errorf("RoleAt called with the wrong arg. Want=%d, got=%d", slot, v.RoleAtArg1)
	}
}

func TestRunOnNextSlotActsAsAttester(t *testing.T) {
	v := &fakeValidator{}
	ctx, cancel := context.WithCancel(context.Background())

	slot := uint64(55)
	ticker := make(chan uint64)
	v.NextSlotRet = ticker
	v.RoleAtRet = pb.ValidatorRole_ATTESTER
	go func() {
		ticker <- slot

		cancel()
	}()

	run(ctx, v)

	if !v.AttestToBlockHeadCalled {
		t.Fatalf("AttestToBlockHead(%d) was not called", slot)
	}
	if v.AttestToBlockHeadArg1 != slot {
		t.Errorf("AttestToBlockHead was called with wrong arg. Want=%d, got=%d", slot, v.AttestToBlockHeadArg1)
	}
}

func TestRunOnNextSlotActsAsProposer(t *testing.T) {
	v := &fakeValidator{}
	ctx, cancel := context.WithCancel(context.Background())

	slot := uint64(55)
	ticker := make(chan uint64)
	v.NextSlotRet = ticker
	v.RoleAtRet = pb.ValidatorRole_PROPOSER
	go func() {
		ticker <- slot

		cancel()
	}()

	run(ctx, v)

	if !v.ProposeBlockCalled {
		t.Fatalf("ProposeBlock(%d) was not called", slot)
	}
	if v.ProposeBlockArg1 != slot {
		t.Errorf("ProposeBlock was called with wrong arg. Want=%d, got=%d", slot, v.AttestToBlockHeadArg1)
	}
}
