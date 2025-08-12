package v1

import (
	"context"
	"testing"

	"emshop/internal/app/user/srv/data/v1/mock"
	metav1 "emshop/pkg/common/meta/v1"
)

func TestUserList(t *testing.T) {
	userSrv := NewUserService(mock.NewUsers())
	result, err := userSrv.List(context.Background(), nil, metav1.ListMeta{})
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if result == nil {
		t.Error("Expected result, got nil")
		return
	}
	
	if result.TotalCount != 2 {
		t.Errorf("Expected TotalCount to be 2, got %d", result.TotalCount)
	}
	
	if len(result.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(result.Items))
	}
}

func TestUserGetByID(t *testing.T) {
	userSrv := NewUserService(mock.NewUsers())
	result, err := userSrv.GetByID(context.Background(), 1)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if result == nil {
		t.Error("Expected result, got nil")
		return
	}
	
	if result.Mobile != "13800138000" {
		t.Errorf("Expected mobile to be '13800138000', got %s", result.Mobile)
	}
}

func TestUserGetByMobile(t *testing.T) {
	userSrv := NewUserService(mock.NewUsers())
	result, err := userSrv.GetByMobile(context.Background(), "13800138000")
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if result == nil {
		t.Error("Expected result, got nil")
		return
	}
	
	if result.Mobile != "13800138000" {
		t.Errorf("Expected mobile to be '13800138000', got %s", result.Mobile)
	}
}