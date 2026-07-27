package dnd

import (
	"testing"

	"github.com/gogpu/ui/geometry"
)

func TestKindFileConstant(t *testing.T) {
	if KindFile != "file" {
		t.Fatalf("KindFile = %q, want %q", KindFile, "file")
	}
}

func TestFilePayload(t *testing.T) {
	paths := []string{"/home/user/a.txt", "/home/user/b.png"}
	payload := FilePayload{Paths: paths}

	if len(payload.Paths) != 2 {
		t.Fatalf("len(payload.Paths) = %d, want 2", len(payload.Paths))
	}
	if payload.Paths[0] != "/home/user/a.txt" {
		t.Fatalf("payload.Paths[0] = %q, want %q", payload.Paths[0], "/home/user/a.txt")
	}
	if payload.Paths[1] != "/home/user/b.png" {
		t.Fatalf("payload.Paths[1] = %q, want %q", payload.Paths[1], "/home/user/b.png")
	}
}

func TestDropExternal_TargetAccepts(t *testing.T) {
	mgr := NewManager()

	target := &mockDropTarget{
		acceptKinds:  []string{KindFile},
		dropAccepted: true,
	}
	mgr.RegisterTarget(target, geometry.NewRect(0, 0, 200, 200))

	data := DragData{Kind: KindFile, Payload: FilePayload{Paths: []string{"/tmp/test.txt"}}}
	ok := mgr.DropExternal(data, 50, 75)

	if !ok {
		t.Fatal("DropExternal returned false, want true")
	}
	if !target.entered {
		t.Fatal("DragEnter not called on target")
	}
	if !target.dropped {
		t.Fatal("Drop not called on target")
	}
	if target.lastDropData.Kind != KindFile {
		t.Fatalf("lastDropData.Kind = %q, want %q", target.lastDropData.Kind, KindFile)
	}
	fp, ok2 := target.lastDropData.Payload.(FilePayload)
	if !ok2 {
		t.Fatal("lastDropData.Payload is not FilePayload")
	}
	if len(fp.Paths) != 1 || fp.Paths[0] != "/tmp/test.txt" {
		t.Fatalf("payload paths = %v, want [/tmp/test.txt]", fp.Paths)
	}
	if target.lastDropPos.X != 50 || target.lastDropPos.Y != 75 {
		t.Fatalf("lastDropPos = (%f, %f), want (50, 75)", target.lastDropPos.X, target.lastDropPos.Y)
	}
}

func TestDropExternal_NoTargetAtPosition(t *testing.T) {
	mgr := NewManager()

	target := &mockDropTarget{acceptKinds: []string{KindFile}, dropAccepted: true}
	mgr.RegisterTarget(target, geometry.NewRect(0, 0, 100, 100))

	data := DragData{Kind: KindFile, Payload: FilePayload{Paths: []string{"/tmp/test.txt"}}}
	ok := mgr.DropExternal(data, 150, 150) // Outside bounds

	if ok {
		t.Fatal("DropExternal returned true for position outside all targets")
	}
}

func TestDropExternal_TargetRejectsKind(t *testing.T) {
	mgr := NewManager()

	target := &mockDropTarget{acceptKinds: []string{"text"}} // Only accepts "text"
	mgr.RegisterTarget(target, geometry.NewRect(0, 0, 200, 200))

	data := DragData{Kind: KindFile, Payload: FilePayload{Paths: []string{"/tmp/test.txt"}}}
	ok := mgr.DropExternal(data, 50, 50)

	if ok {
		t.Fatal("DropExternal returned true when target rejects kind")
	}
}

func TestDropExternal_NoTargets(t *testing.T) {
	mgr := NewManager()

	data := DragData{Kind: KindFile, Payload: FilePayload{Paths: []string{"/tmp/test.txt"}}}
	ok := mgr.DropExternal(data, 50, 50)

	if ok {
		t.Fatal("DropExternal returned true with no registered targets")
	}
}

func TestDropExternal_TargetDropReturnsFalse(t *testing.T) {
	mgr := NewManager()

	target := &mockDropTarget{
		acceptKinds:  []string{KindFile},
		dropAccepted: false, // Target explicitly rejects the drop
	}
	mgr.RegisterTarget(target, geometry.NewRect(0, 0, 200, 200))

	data := DragData{Kind: KindFile, Payload: FilePayload{Paths: []string{"/tmp/test.txt"}}}
	ok := mgr.DropExternal(data, 50, 50)

	if ok {
		t.Fatal("DropExternal returned true when Drop() returned false")
	}
	if !target.entered {
		t.Fatal("DragEnter not called even though CanAccept returned true")
	}
}
