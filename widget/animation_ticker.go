package widget

// AnimationTicker is implemented by widgets whose animations affect layout
// (e.g., Collapsible height, Transition size). The framework calls
// TickAnimation on every frame BEFORE the layout pass, following the
// Flutter pattern: handleBeginFrame (animate) → handleDrawFrame (layout).
//
// Layout-affecting animations must tick here, not inside Layout() or Draw(),
// so that Layout remains a pure function of (constraints + widget state).
// This invariant is required for RelayoutBoundary (ADR-032 Phase 5).
//
// Paint-only animations (spinner rotation, cursor blink) do NOT implement
// this interface — they use ScheduleAnimationFrame + SetNeedsRedraw instead.
type AnimationTicker interface {
	TickAnimation(ctx Context)
}
