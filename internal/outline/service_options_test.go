package outline

import (
	"testing"
)

func TestWithDeleter_SetsDeleter(t *testing.T) {
	d := &fakeFileDeleter{}
	svc := NewOutlineService(nil, nil, &mockLocker{}, nil, WithDeleter(d))

	if svc.deleter != d {
		t.Error("WithDeleter did not set deleter field")
	}
}

func TestWithRenamer_SetsRenamer(t *testing.T) {
	r := &fakeFileRenamer{}
	svc := NewOutlineService(nil, nil, &mockLocker{}, nil, WithRenamer(r))

	if svc.renamer != r {
		t.Error("WithRenamer did not set renamer field")
	}
}

func TestWithContentReader_SetsContentReader(t *testing.T) {
	cr := &fakeContentReader{}
	svc := NewOutlineService(nil, nil, &mockLocker{}, nil, WithContentReader(cr))

	if svc.contentReader != cr {
		t.Error("WithContentReader did not set contentReader field")
	}
}

func TestWithSlugifier_SetsSlugifier(t *testing.T) {
	sl := &stubSlugifier{slug: "test-slug"}
	svc := NewOutlineService(nil, nil, &mockLocker{}, nil, WithSlugifier(sl))

	if svc.slugifier != sl {
		t.Error("WithSlugifier did not set slugifier field")
	}
}

func TestWithFrontmatterHandler_SetsFMHandler(t *testing.T) {
	fh := &stubFrontmatterHandler{}
	svc := NewOutlineService(nil, nil, &mockLocker{}, nil, WithFrontmatterHandler(fh))

	if svc.fmHandler != fh {
		t.Error("WithFrontmatterHandler did not set fmHandler field")
	}
}

func TestNewOutlineService_DefaultsWithoutOptions(t *testing.T) {
	svc := NewOutlineService(nil, nil, &mockLocker{}, nil)

	if svc.builder == nil {
		t.Error("builder should have a default value")
	}
}

func TestNewOutlineService_MultipleOptions(t *testing.T) {
	d := &fakeFileDeleter{}
	r := &fakeFileRenamer{}
	cr := &fakeContentReader{}

	svc := NewOutlineService(nil, nil, &mockLocker{}, nil,
		WithDeleter(d),
		WithRenamer(r),
		WithContentReader(cr),
	)

	if svc.deleter != d {
		t.Error("WithDeleter not applied in multi-option call")
	}
	if svc.renamer != r {
		t.Error("WithRenamer not applied in multi-option call")
	}
	if svc.contentReader != cr {
		t.Error("WithContentReader not applied in multi-option call")
	}
}

// stubSlugifier is a test double for Slugifier that returns a fixed slug.
type stubSlugifier struct {
	slug string
}

func (s *stubSlugifier) Slug(_ string) string { return s.slug }

// stubFrontmatterHandler is a test double for FrontmatterHandler.
type stubFrontmatterHandler struct{}

func (s *stubFrontmatterHandler) GetTitle(input string) (string, error)           { return "", nil }
func (s *stubFrontmatterHandler) SetTitle(input, newTitle string) (string, error) { return "", nil }
func (s *stubFrontmatterHandler) EncodeYAMLValue(str string) string               { return str }
func (s *stubFrontmatterHandler) Serialize(fm, body string) string                { return fm }
