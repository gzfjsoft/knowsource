package logic

import (
	"context"
	"testing"
)

const AdminJWT = ""

func TestListVhosts(t *testing.T) {
	result, err := AdminListVhosts(context.Background(), AdminJWT, &ListVhostsRequest{
		Host: "dev.xxx.cn",
	})
	if err != nil {
		t.Errorf("AdminListVhosts failed: %v", err)
	}
	t.Logf("AdminListVhosts result: %v", result)

	result2, err := ListVhosts(context.Background(), AdminJWT, &ListVhostsRequest{
		Host: "dev.xxx.cn",
	})
	if err != nil {
		t.Errorf("AdminListVhosts failed: %v", err)
	}
	t.Logf("ListVhosts result: %v", result2)

}
