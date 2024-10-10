## 0.6.0 (October 10, 2024)

FEATURES:

* vault-secrets: Add support for creating rotating and dynamic secrets interactively [[GH-167](https://github.com/hashicorp/hcp/issues/167)]
* vault-secrets: Add support for managing vault-secrets integrations and rotating/dynamic secrets [[GH-176](https://github.com/hashicorp/hcp/issues/176)]

## 0.5.0 (September 3, 2024)

FEATURES:

* vault-secrets: CRUD operations for vault-secrets gateway pools [[GH-131](https://github.com/hashicorp/hcp/issues/131)]

IMPROVEMENTS:

* Support loading all variables from Waypoint server in Waypoint agent CLI [[GH-148](https://github.com/hashicorp/hcp/issues/148)]
* vault-secrets: Enhances dynamic secrets output [[GH-129](https://github.com/hashicorp/hcp/issues/129)]
* vault-secrets: adding list gateway pools gateways command to vault-secrets gateway-pools [[GH-133](https://github.com/hashicorp/hcp/issues/133)]
* vault-secrets: improving vault-secrets gateway-pools read to include associated integrations [[GH-132](https://github.com/hashicorp/hcp/issues/132)]
* vault-secrets: storing credentials and config files for gateway create [[GH-135](https://github.com/hashicorp/hcp/issues/135)]
* waypoint: Add support for creating/updating Waypoint Templates with Variables [[GH-130](https://github.com/hashicorp/hcp/issues/130)]
* waypoint: Remove module version flag from templates and add-on definitions commands. Remove module source from update commands. [[GH-118](https://github.com/hashicorp/hcp/issues/118)]

BUG FIXES:

* include all secrets from paginated respoonses when invoking `hcp vs run` command [[GH-147](https://github.com/hashicorp/hcp/issues/147)]
* security: address vulnerabilities for docker image CVE-2024-7264 (curl) and CVE-2024-43374 (vim) [[GH-151](https://github.com/hashicorp/hcp/issues/151)]
* security: address vulnerability for docker image CVE-2024-43790 / CVE-2024-43802 (vim) [[GH-152](https://github.com/hashicorp/hcp/issues/152)]
* vault-secrets: issue an error if formatted secret names collide during a run command [[GH-127](https://github.com/hashicorp/hcp/issues/127)]

## 0.4.0 (June 25, 2024)

FEATURES:

* vault-secrets: Adds support for dynamic secrets to the `secrets open` and `run` commands. [[GH-119](https://github.com/hashicorp/hcp/issues/119)]

IMPROVEMENTS:

* Run auto-detection of organization ID even if project ID is configured in the profile. [[GH-122](https://github.com/hashicorp/hcp/issues/122)]
* Turn actions and agents sub-commands in waypoint back on [[GH-120](https://github.com/hashicorp/hcp/issues/120)]
* profile: Add a core/quiet property which allows disabling prompting in the profile. [[GH-121](https://github.com/hashicorp/hcp/issues/121)]
* vault-secrets: Adds secret type to the `secrets read` and `secrets list` output. [[GH-119](https://github.com/hashicorp/hcp/issues/119)]

## 0.3.0 (June 11, 2024)

FEATURES:

* iam: Adds `read-policy`, `set-policy`, `add-binding`, and `delete-binding` subcommands to `hcp iam groups iam` which allow the ability to manage an IAM policy on a group.
- `read-policy` Reads an IAM policy for a specified group.
- `set-policy` Sets an IAM policy for a group using a JSON file.
- `add-binding` Adds a single role binding to a user principal.
- `delete-binding` Removes a single role binding from a user principal. [[GH-113](https://github.com/hashicorp/hcp/issues/113)]
* vault-secrets: Add `vault-secrets` CLI for managing Vault Secrets resources. [[GH-105](https://github.com/hashicorp/hcp/issues/105)]

## 0.2.0 (May 31, 2024)

BREAKING CHANGES:

* Removing waypoint actions and agents sub-command [[GH-100](https://github.com/hashicorp/hcp/issues/100)]
* Renames the -application-name flag for creating waypoint add-ons to just -app [[GH-77](https://github.com/hashicorp/hcp/issues/77)]

FEATURES:

* iam groups: Add update group command. [[GH-41](https://github.com/hashicorp/hcp/issues/41)]
* waypoint: Add `hcp waypoint add-ons definitions` CLI for managing HCP Waypoint add-on definitions. [[GH-44](https://github.com/hashicorp/hcp/issues/44)]
* waypoint: Add `hcp waypoint add-ons` CLI for managing HCP Waypoint add-ons. [[GH-52](https://github.com/hashicorp/hcp/issues/52)]
* waypoint: Add `waypoint applications` CLI for managing HCP Waypoint applications. [[GH-48](https://github.com/hashicorp/hcp/issues/48)]
* waypoint: Add `waypoint templates` CLI for managing HCP Waypoint templates. [[GH-40](https://github.com/hashicorp/hcp/issues/40)]

IMPROVEMENTS:

* A better error message is now returned if a user attempts to read a Waypoint TFC Config when one has not been set [[GH-63](https://github.com/hashicorp/hcp/issues/63)]
* Adds opinionated sorting to the listing of iam roles [[GH-67](https://github.com/hashicorp/hcp/issues/67)]
* Detect new versions of the CLI and prompt for update. [[GH-91](https://github.com/hashicorp/hcp/issues/91)]
* auth: If authenticating as a service principal, automatically populate the profile with the organization and project ID. This allows using the CLI without instantiating the profile. [[GH-46](https://github.com/hashicorp/hcp/issues/46)]
* format: Improved table output when length exceeds terminal width. [[GH-90](https://github.com/hashicorp/hcp/issues/90)]
* waypoint: Rename HCP Waypoint action config command group. [[GH-61](https://github.com/hashicorp/hcp/issues/61)]
* waypoint: change template list format to use a table [[GH-60](https://github.com/hashicorp/hcp/issues/60)]

BUG FIXES:

* Fix rare panic that could occur on authentication error when running a command that had quoted arguments. [[GH-98](https://github.com/hashicorp/hcp/issues/98)]
* profile: Some profile and profiles commands required authentication unnecessarily. [[GH-47](https://github.com/hashicorp/hcp/issues/47)]
* waypoint: Fix table output for agent group list command [[GH-59](https://github.com/hashicorp/hcp/issues/59)]
* waypoint: Update API client to use correct field names for action run id. [[GH-57](https://github.com/hashicorp/hcp/issues/57)]
