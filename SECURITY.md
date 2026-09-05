# Security Policy

## Template security model

Go42 is a project template, not a managed service or runtime distribution. A
CLI tool fetches a selected Go42 release to initialize an independent repository
that is operated in the user's private infrastructure.

Each release is a bootstrap snapshot. Go42 does not automatically update or
backport security fixes to repositories created from that snapshot. Users are
responsible for the security of their generated repository and deployment,
including reviewing and updating dependencies, container images, workflows,
configuration, infrastructure, and application code.

New Go42 releases may contain security improvements. Users decide whether and
how to apply those changes to an existing project.

## Reporting a vulnerability

Do not disclose security vulnerabilities through public GitHub issues,
discussions, or pull requests.

Use [GitHub private vulnerability reporting](https://github.com/hasansino/go42/security/advisories/new)
to report a vulnerability in the Go42 template or its distribution tooling.
Include the affected version, potential impact, reproduction steps, and any
known mitigations when possible.

Vulnerabilities introduced by downstream customization or private infrastructure
are the responsibility of that project's owner and should be handled within that
project's security process.

We will acknowledge a vulnerability report within 7 days. We will coordinate
investigation, remediation, and public disclosure with the reporter. Disclosure
should occur only after a fix is available or on another mutually agreed timeline.
