package data

import (
    _ "embed"

    "git.ierislabs.dev/update-link/utils"
)

//go:embed distribution.conf
var domainFile string
var DistributionDomains = utils.ParseFileLines(domainFile)
