# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

container {
	dependencies = true
	alpine_secdb = true
	secrets      = true

  triage {
    suppress {
      vulnerabilities = [
        // curl is used in "Print access token" command as an example and added
        // for convinience in the Docker image to support this example.
        "CVE-2024-7264", // curl@8.9.0-r0; no new image with fix; RE: https://security.alpinelinux.org/vuln/CVE-2024-7264
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
