# Executive Plan

Date: 2026-03-07
Source: `AUDIT.md`

## Executive Summary

Admin frontend has a solid operational core, but several configuration and experimentation areas are still mock-level UI. The immediate goal is to remove false signals of completeness, complete the highest-value admin workflows, and then finish secondary analytical/optimization surfaces.

Important update: the backend foundation for advanced bandit functionality is no longer the blocker. Advanced bandit endpoints are now live on the local Docker runtime, persist objective configuration, expose objective/window/metrics data, and have a reusable smoke test script.

## Priority Framework

- **P0** — Must fix now to avoid misleading users or blocking core admin operations
- **P1** — Important business workflows that should follow immediately after P0
- **P2** — Valuable enhancements and lower-risk completion work

## P0 — Stop the UX debt and enable core admin control

### Goals
- Prevent users from treating unfinished screens as production-ready
- Complete the most important admin configuration surfaces
- Reduce operational confusion between real and mock functionality

### Scope
1. **Finish `Platform Settings` end-to-end**
   - Load saved settings from backend
   - Save updates for general settings
   - Save integration credentials/config
   - Save notification preferences
   - Save security settings where safe
   - Add validation, loading states, success/error feedback

2. **Mark unfinished sidebar items as `comingSoon` or temporarily hide them**
   - `Matomo Analytics`
   - `Platform Settings`
   - `Dunning`
   - `Winback`
   - `Pricing Tiers`
   - `A/B Tests`
   - `Studio`
   - `Bandit Model`
   - `Delayed Feedback`
   - `Sliding Window`
   - `Multi-Objective`

3. **Decide product direction for standalone `Dunning` page**
   - Either remove it from sidebar
   - Or wire it to the same live data/domain model already shown in `Revenue Ops`

### Success Criteria
- Unfinished pages are clearly labeled or hidden
- `Platform Settings` becomes a real working admin surface
- `Dunning` no longer exists as a misleading mock page

## P1 — Complete monetization operations

### Goals
- Enable real admin action for pricing and recovery workflows
- Turn revenue-adjacent mock screens into usable tools

### Scope
1. **Finish `Pricing Tiers`**
   - List plans from backend
   - Create/edit/deactivate pricing plans
   - Persist trial, grace, interval, currency fields
   - Add validation and mutation feedback

2. **Finish `Winback`**
   - Load campaigns from backend
   - Create and edit campaign configuration
   - Launch/pause/archive campaigns
   - Add targeting and expiration persistence

3. **Stabilize partial gaps in working operational screens**
   - Move `Subscriptions` sorting to backend API
   - Move `Transactions` sorting to backend API
   - Review missing filter/sort parity across list screens

### Success Criteria
- Pricing management supports real CRUD/admin actions
- Winback campaigns can be managed without code changes
- Operational tables no longer rely on temporary frontend sorting hacks

## P2 — Finish experimentation and analytics surfaces

### Goals
- Complete strategic optimization tooling after core revenue/admin workflows are stable
- Replace static experiment pages with real data and actions
- Reuse the already-working backend bandit API instead of building a second parallel integration path

### Scope
1. **Finish `Experiments` overview**
   - Load real experiments
   - Support draft/running/completed lifecycle
   - Wire create, launch, stop, archive actions

2. **Finish `Experiment Studio`**
   - Draft creation/editing
   - Variant configuration persistence
   - Targeting rules persistence
   - Launch flow

3. **Finish advanced experiment pages**
   - `Bandit Model`
   - `Delayed Feedback`
   - `Sliding Window`
   - `Multi-Objective`

   Backend foundation already available and verified:
   - objective scores/config endpoints are live
   - sliding-window endpoints are live
   - delayed reward read endpoints are live
   - metrics endpoint is live
   - smoke script exists at `scripts/test_advanced_bandit_endpoints.sh`

4. **Finish `Matomo Analytics`**
   - Decide embed vs API integration approach
   - Load real KPIs
   - Replace placeholder iframe state
   - Wire external/open actions if needed

### Success Criteria
- Experiment screens are no longer static mock UI
- Existing advanced bandit backend endpoints are consumed directly by frontend pages/actions
- Analytics page reflects real Matomo integration or is intentionally deferred/labeled
- Product optimization tooling is credible for internal use

## Recommended Delivery Order

1. P0 — truth in navigation + real settings + dunning decision
2. P1 — pricing + winback + API cleanup for ops tables
3. P2 — experiments suite + Matomo

Within P2, the recommended experiment order is now:

1. `Experiments` overview
2. `Experiment Studio`
3. frontend wiring for advanced bandit pages to existing backend endpoints
4. `Matomo`

## Risks

- Leaving mock pages visible creates trust damage for internal users
- Building experiments before settings/pricing may delay higher-value admin workflows
- Duplicating `Dunning` concepts outside `Revenue Ops` may create inconsistent behavior
- Rebuilding advanced bandit backend contracts on the frontend side would waste time, because the local runtime already demonstrates a usable contract

## Immediate Next Actions

1. Update sidebar status for unfinished pages
2. Define backend contract for `Platform Settings`
3. Decide whether `Dunning` remains standalone
4. Start implementation of `Platform Settings` as the first P0 delivery
5. After P0/P1, wire `experiments/*` frontend pages to the verified advanced bandit API instead of leaving them on mock arrays

