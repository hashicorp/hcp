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

## PCI review checklist

<!-- heimdall_github_prtemplate:grc-pci_dss-2024-01-05 -->

- [ ] If applicable, I've documented a plan to revert these changes if they require more than reverting the pull request.

- [ ] If applicable, I've worked with GRC to document the impact of any changes to security controls.

  Examples of changes to controls include access controls, encryption, logging, etc.

- [ ] If applicable, I've worked with GRC to ensure compliance due to a significant change to the in-scope PCI environment.

  Examples include changes to operating systems, ports, protocols, services, cryptography-related components, PII processing code, etc.