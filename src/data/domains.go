package data

import (
    _ "embed"

    "ieris19.com/podman-updater/utils"
)

//go:embed distribution.conf
var domainFile string
var DistributionDomains = utils.ParseFileLines(domainFile)
