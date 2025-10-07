# Copyright IBM Corp. 2024, 2025
# SPDX-License-Identifier: MPL-2.0

container {
	dependencies = true
	alpine_secdb = true
	secrets      = true
	triage {
		suppress {
			// there is currently no release available for jq
			// https://security.alpinelinux.org/vuln/CVE-2024-53427
			vulnerabilities = [
				"CVE-2024-53427",
				"CVE-2025-46394",
				"CVE-2024-58251",
				"CVE-2024-23337",
				"CVE-2025-48060",
			]
		}
  	}
}

binary {
	secrets      = true
	go_modules   = true
	osv          = true
	oss_index    = false
	nvd          = false
}
