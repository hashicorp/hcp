### :hammer_and_wrench:  Description

<!-- What changed?
<!-- Why was it changed? -->
<!-- How does it affect end-user behavior? -->

### :link:  Additional Link

<!-- Any additional link to understand the context of the change. -->

### :building_construction:  Local Testing

<!-- List steps to test your change on a local environment. -->

### :+1:  Checklist

- [ ] The PR has a descriptive title.
- [ ] Input validation updated
- [ ] Unit tests updated
- [ ] Documentation updated
- [ ] Major architecture changes have a corresponding RFC
- [ ] Tests added if applicable
- [ ] CHANGELOG entry added or label 'pr/no-changelog' added to PR
  > Run `CHANGELOG_PR=<PR number> make changelog/new-entry` for guidance
  > in authoring a changelog entry, and commit the resulting file, which should
  > have a name matching your PR number. Entries should use imperative present
  > tense (e.g. Add support for...)

<!-- heimdall_github_prtemplate:grc-pci_dss-2024-01-05 -->
## Rollback Plan

If a change needs to be reverted, we will roll out an update to the code within 7 days.

## Changes to Security Controls

Are there any changes to security controls (access controls, encryption, logging) in this pull request? If so, explain.