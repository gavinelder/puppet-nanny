# Puppet-Nanny

[![Go Report Card](https://goreportcard.com/badge/github.com/gavinelder/puppet-nanny)](https://goreportcard.com/report/github.com/gavinelder/puppet-nanny)

:warning: The following is hacky

PRs welcome.

## Overview

Long term this application is going to be a Cross OS compliant tool for managing puppet runs as a seperate binary which does not conflict with the main puppet system.

Right now the Script needs some tidy up but it can do the following:

- Check for appropriate permissions before proceeding
- Run in an infinite loop for use as a service
- Check if puppet has been disabled and remove the lock file
- Check the age of the lock file and remove if it's stale or abort run and try again later.

Long term I want to:

- Refactor it.
- Implement logging (Avoided currently with view of using cross OS compliant logging lib)
- Bootstrap itself
- Allow passing of flags for Environment , Run time or other...
