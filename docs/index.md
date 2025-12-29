# Secret Santa

![Secret Santa](./images/secret-santa.webp)

**Secret Santa** is a Kubernetes operator for generating secrets with templates and storing them in multiple destinations.

## What is Secret Santa?

Secret Santa is a Kubernetes operator that generates secrets using configurable templates and stores them across multiple destinations including Kubernetes secrets, AWS Secrets Manager, AWS Parameter Store, and GCP Secret Manager. It provides a declarative way to manage secret generation and distribution in cloud-native environments.

### Key Features

- **Multiple Storage Destinations**: Kubernetes secrets, AWS Secrets Manager, AWS Parameter Store, GCP Secret Manager
- **Template Engine**: Go templates with crypto, random, and TLS generators
- **Create-Once Semantics**: Secrets generated once and never modified
- **Cloud Integration**: Native AWS and GCP authentication support
- **Dry-Run Mode**: Validate templates and preview masked output without creating secrets
- **Automatic Metadata**: Built-in traceability and observability

### Where to get started

To get started, please read through the [core concepts](introduction/concepts.md) to understand the architecture and design principles. Next, follow our [installation guide](guides/installation.md) to install Secret Santa in your cluster, then try the [quick start guide](guides/quick-start.md) for your first secret.

For complete examples, check out our [examples](examples/) section, and for detailed configuration options, see our [guides](guides/).

### How to get involved

This project is part of the LogicIQ ecosystem and we welcome contributions:

- [GitHub Issues](https://github.com/LogicIQ/secret-santa/issues)
- [Contributing Guide](contributing/process.md)
- [LogicIQ Website](https://logiciq.ca/secret-santa)

### Architecture Overview

Secret Santa follows a simple but powerful architecture:

1. **SecretSanta Custom Resource**: Defines what secrets to generate and where to store them
2. **Template Engine**: Processes Go templates with generated values
3. **Generators**: Create cryptographic material, passwords, and other secret data
4. **Media Providers**: Store secrets in various destinations (K8s, AWS, GCP)

The operator continuously reconciles SecretSanta resources, ensuring secrets exist and are properly distributed across configured destinations.