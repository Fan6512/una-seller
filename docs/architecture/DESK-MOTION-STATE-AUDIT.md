# Desktop motion system — state-change audit and acceptance

User goal: Apple-like clarity, restraint and interruption, not decorative animation. This PR is the first implementation candidate, not a measured final quality claim. No dependency or API change.

## Audit by change, not by widget

| State change | Existing issue / approach | Implemented |
| --- | --- | --- |
| Press / release, hover, focus, disabled | scattered 120–220ms easing, inconsistent feedback | unified 100/160ms easing, small action-only scale, stable focus and disabled styles |
| Selection / status / theme | tabs/labels jumped or changed font metrics | selected/active/status color transitions, stable tab/counter type metrics |
| Appear / switch content | login, detail, editing language, messages and empty/error states abruptly appeared | bounded entry 160–200ms, no per-row animations; existing conditional mounting behavior is unchanged |
| Modal open / close | abrupt native dialogs | 220ms native entrance and backdrop opacity; immediate native close intentionally preserves focus and clears private data; no delayed exit |
| Image/content replaces | gallery switches before image ready | only animate successful main-image load and explicit preview, no extra prefetch, fixed gallery geometry; existing errors remain visible |
| Numbers and progress | layout shifts / noisy repeated updates | exact values, tabular numbers; low-frequency selection/page/gallery counters get cancellable opacity cues; typing, financial values, upload progress and countdown never interpolate |
| Expand / collapse / resize / move | native details and textarea resizing | no height/width animation, direct interaction; no invented drag interaction or FLIP on list changes |
| Loading → completed / error | inconsistent notices | short feedback entrance; genuine spinner only; no endless success pulse |
| Reduce motion | prior 0.01ms infinite spin risk | all CSS motion off, WAAPI cancellation on preference change, visible static spinner and normal text |
| Long content | avoid costly compositing | no transition:all, no per-row entrance, no permanent will-change, no layout property transitions |

Reviewed App, Login, desktop updater, account/store/title settings, orders/detail/billing, published browser/editor, collection/product-box, workbench/gallery/variant and AI/attribute states via source. Not all dynamic content is fully animated; exits, high-frequency text, long layout shifts, native select/checkbox remain intentionally direct. Switching primary sections can unmount pages; hiding the active order/listing page behind Settings preserves its existing mounted state. The current baseline explicitly declares color-scheme: light; no true dark mode exists. Motion is color-independent, but dark-theme support/verification cannot be claimed until a separate semantic theme migration. Existing fixed light colors are unchanged.

## Validation (mandatory before quality sign-off)

- CI static motion invariant test and TSC/Vite. Existing workbench tests still pass.
- GitHub Actions Windows test build only; actual WebView2 at 780, 1024, 1440 and wide 1920 widths; light/dark and OS reduce motion. Test app shell, repeated order/details, language switches, image load failure/cache/rapid replacement, product selection, upload/error, resizing, keyboard Esc and focus return. Confirm no hidden-page interactions, clipping, flash or jumping; no stale image/data.
- Capture 5-second Performance traces at idle, rapid interaction (at least 10 cycles), large list scroll, modal open/close, image selection; compare baseline and candidate on same Windows/WebView2/device, 100+ rows/variants where supported. Record refresh rate, frame duration p50/p95, dropped frame ratio, long tasks >50ms, layout count, GPU memory if available, reduced-motion trace. Aim at refresh-budget p95 (16.7ms for 60Hz), no extra recurring layout/long tasks, no material regression; report actual measurements rather than assert full-frame. On slow devices remove expensive animation first.
- Accessibility: native controls, keyboard, screen-reader status announcements unchanged; motion never gates input/authorization or save. Reduced motion mid-flight cancels; no fake numeric transitions. Verify native modal close and rapid reopening. 
- Real Windows measurements and screenshots NOT_RUN in this PR creation environment; no release or deployment. Further state-specific refinement after real traces, do not call this final Apple-level sign-off.
