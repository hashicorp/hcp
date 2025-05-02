## 0.10.0 (May 2, 2025)

BUG FIXES:

* Fix a bug preventing the usage of pretty printing [[GH-225](https://github.com/hashicorp/hcp/issues/225)]
* waypoint: Fix a bug preventing Waypoint actions from being created or updated. [[GH-224](https://github.com/hashicorp/hcp/issues/224)]

## 0.9.0 (March 18, 2025)

BREAKING CHANGES:

* waypoint: Require no-code module ID for creating templates and add-on defintiions. [[GH-198](https://github.com/hashicorp/hcp/issues/198)]

BUG FIXES:

* security: address vulnerability GO-2025-3487 and GO-2025-3488. [[GH-219](https://github.com/hashicorp/hcp/issues/219)]
* vault-secrets: corrected error output for postgres integration creation, MongoDB Atlas -> Postgres. [[GH-203](https://github.com/hashicorp/hcp/issues/203)]
* vault-secrets: corrected json output for gateways, GatewayPool -> gateway_pool [[GH-201](https://github.com/hashicorp/hcp/issues/201)]
* vault-secrets: fixed list integrations request to pull in all integrations when no parameters specified [[GH-222](https://github.com/hashicorp/hcp/issues/222)]
* waypoint: Fix panic listing Waypoint agent groups when none exist. [[GH-220](https://github.com/hashicorp/hcp/issues/220)]

## 0.8.0 (December 04, 2024)

FEATURES:

* hvs: support postgres rotating secret CRUDL [[GH-192](https://github.com/hashicorp/hcp/issues/192)]
* vault-secrets: Adds support for creating / updating / reading / deleting Postgres integrations. [[GH-189](https://github.com/hashicorp/hcp/issues/189)]

IMPROVEMENTS:

* HVS: Use stable/2023-11-28 api [[GH-197](https://github.com/hashicorp/hcp/issues/197)]
* vault-secrets: adding Gateway Pool Resource ID to the output of gateway commands. [[GH-195](https://github.com/hashicorp/hcp/issues/195)]

## 0.7.0 (November 04, 2024)

BREAKING CHANGES:

* waypoint: Remove type field from variable options config file. [[GH-174](https://github.com/hashicorp/hcp/issues/174)]

FEATURES:

* waypoint: Add ability for variable options on an add-on definition to be configured with an HCL file. [[GH-170](https://github.com/hashicorp/hcp/issues/170)]
* waypoint: Add flags to make execution mode configurable for Waypoint templates and add-on definitions. [[GH-163](https://github.com/hashicorp/hcp/issues/163)]
* waypoint: Support setting variables when creating apps. [[GH-172](https://github.com/hashicorp/hcp/issues/172)]
* waypoint: Support setting variables when installing add-ons. [[GH-173](https://github.com/hashicorp/hcp/issues/173)]

IMPROVEMENTS:

* Waypoint: Action sequence numbers are now reported on agent run completion [[GH-165](https://github.com/hashicorp/hcp/issues/165)]
* vault-secrets: Update vault-secrets rotating secrets from `secret_name` to `name` usage [[GH-184](https://github.com/hashicorp/hcp/issues/184)]

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
