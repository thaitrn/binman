package tui

import (
	"strings"
	"testing"

	"github.com/thaitrn/binman/internal/app"
	"github.com/thaitrn/binman/internal/apps"
)

func TestPickerModel_ViewRenders(t *testing.T) {
	entries := []apps.Entry{
		{App: &app.App{Name: "Alpha", BundleID: "com.x.alpha", Path: "/Applications/Alpha.app"}, Size: 1024},
		{App: &app.App{Name: "Beta", BundleID: "com.x.beta", Path: "/Applications/Beta.app"}, Size: 2048, Protected: true},
	}
	m := newPickerModel(entries)
	m.list.SetWidth(100)
	out := m.View()
	for _, want := range []string{"Select an app", "Alpha", "Beta", "system"} {
		if !strings.Contains(out, want) {
			t.Errorf("picker View() missing %q\n---\n%s", want, out)
		}
	}
}

func TestPickApp_Empty(t *testing.T) {
	got, ok, err := PickApp(nil)
	if err != nil || ok || got != nil {
		t.Errorf("PickApp(nil) = %+v ok=%v err=%v, want nil/false/nil", got, ok, err)
	}
}
