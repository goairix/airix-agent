package model_test

import (
	"sort"
	"testing"

	"github.com/dysodeng/app/internal/domain/agent/session/model"
	"github.com/dysodeng/app/internal/domain/agent/session/valueobject"
)

func TestSession_Validate(t *testing.T) {
	s := &model.Session{Status: valueobject.SessionStatusRunning}
	if err := s.Validate(); err == nil {
		t.Error("session without AgentID should fail")
	}
}

func TestMessageItemByOrder_Sort(t *testing.T) {
	items := model.ByOrder{
		{SortOrder: 2},
		{SortOrder: 0},
		{SortOrder: 1},
	}
	sort.Sort(items)
	if items[0].SortOrder != 0 || items[2].SortOrder != 2 {
		t.Error("sort order incorrect")
	}
}
