schema = 1
artifacts {
  zip = [
    "hcp_${version}_darwin_amd64.zip",
    "hcp_${version}_darwin_arm64.zip",
    "hcp_${version}_linux_386.zip",
    "hcp_${version}_linux_amd64.zip",
    "hcp_${version}_linux_arm.zip",
    "hcp_${version}_linux_arm64.zip",
    "hcp_${version}_windows_386.zip",
    "hcp_${version}_windows_amd64.zip",
    "hcp_${version}_windows_arm.zip",
    "hcp_${version}_windows_arm64.zip",
  ]
  rpm = [
    "hcp-${version_linux}-1.aarch64.rpm",
    "hcp-${version_linux}-1.armv7hl.rpm",
    "hcp-${version_linux}-1.i386.rpm",
    "hcp-${version_linux}-1.x86_64.rpm",
  ]
  deb = [
    "hcp_${version_linux}-1_amd64.deb",
    "hcp_${version_linux}-1_arm64.deb",
    "hcp_${version_linux}-1_armhf.deb",
    "hcp_${version_linux}-1_i386.deb",
  ]
  container = [
    "hcp_release_linux_amd64_${version}_${commit_sha}.docker.dev.tar",
    "hcp_release_linux_amd64_${version}_${commit_sha}.docker.tar",
    "hcp_release_linux_arm64_${version}_${commit_sha}.docker.dev.tar",
    "hcp_release_linux_arm64_${version}_${commit_sha}.docker.tar",
    "hcp_release_linux_arm_${version}_${commit_sha}.docker.dev.tar",
    "hcp_release_linux_arm_${version}_${commit_sha}.docker.tar",
  ]
}
