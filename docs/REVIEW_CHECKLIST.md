# Review Checklist

## Architecture
- [ ] Architecture aligned with hexagonal design
- [ ] Security implications reviewed
- [ ] Performance impact considered
- [ ] Tests added or updated
- [ ] Documentation updated
- [ ] Metrics added (when applicable)
- [ ] Tracing added (when applicable)

## Quality gate
- [ ] `make quality` passes locally
- [ ] Global coverage threshold respected
- [ ] Changed-code coverage threshold respected on pull requests
- [ ] Regression test added for bug fixes
- [ ] No thresholds weakened to make CI pass

See also [quality/checklists/quality-review.md](../quality/checklists/quality-review.md).
