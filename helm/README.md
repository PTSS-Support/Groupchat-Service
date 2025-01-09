# Helm Charts for Quarkus Microservices

This directory contains the Helm charts used to deploy Go microservices on OpenShift. These charts are designed to be a base template that can be used and modified for future microservices.

## Structure

```
helm/
├── Chart.yaml             # Chart metadata
├── values.yaml           # Default values (base configuration)
├── values-dev.yaml      # Development environment values
├── values-prod.yaml     # Production environment values
└── templates/           # Kubernetes/OpenShift resource templates
    ├── deployment.yaml  # Main application deployment
    ├── service.yaml    # Service for internal communication
    └── route.yaml      # OpenShift route for external access
```

## Configuration

### Base Configuration (values.yaml)

The base configuration includes:
- Image settings (registry, repository, tag)
- Resource limits and requests
- Port configurations
- Route settings
- Security context settings

Key configurations that can be customized:
```yaml
image:
  registry: ghcr.io
  repository: ""  # Set via workflow
  tag: latest    # Set via workflow
  pullPolicy: IfNotPresent

resources:
  limits:
    cpu: 300m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 256Mi
```

### Environment-Specific Values

- `values-dev.yaml`: Development environment settings
- `values-prod.yaml`: Production environment settings

## Deployment

The charts are deployed via GitHub Actions workflows. The deployment process:
1. Builds the application
2. Creates a Docker image
3. Pushes to GitHub Container Registry
4. Deploys to OpenShift using Helm

### Required Secrets

The following secrets must be configured in GitHub:
- `OSC_SERVER`: OpenShift server URL
- `OSC_DEV_TOKEN`: OpenShift development login token
- `OSC_PROD_TOKEN`: OpenShift production login token

### Required Variables

The following variables must be configured in GitHub:
- `OSC_DEV_PROJECT`: OpenShift development project name
- `OSC_PROD_PROJECT`: OpenShift production project name
- `APPS_DOMAIN`: Your apps domain (e.g., `apps.inholland.hcs-lab.nl`)

## Security

The charts include basic security settings:
- Non-root user execution
- Dropped capabilities
- TLS configuration for routes

## Customization

When using this template for a new microservice:

1. Update the `Chart.yaml` with your service information
2. Modify resource limits in `values.yaml` based on your needs
3. Configure environment-specific settings in `values-dev.yaml` and `values-prod.yaml`
4. Add any additional templates needed for your service

## Best Practices

1. **Resource Management**
    - Always specify resource requests and limits
    - Monitor resource usage and adjust accordingly

2. **Security**
    - Keep the security context settings
    - Enable TLS for routes
    - Follow the principle of least privilege

3. **Configuration**
    - Use environment-specific value files
    - Keep sensitive information in secrets
    - Document all custom values

4. **Maintenance**
    - Regularly update the chart version
    - Keep dependencies up to date
    - Test changes in development before production

## Adding New Features

Common additions for microservices:
- ConfigMaps for configuration
- Secrets for sensitive data
- Additional services for new endpoints
- Network policies for security
- Volume mounts for persistent data

## Troubleshooting

Common issues and solutions:
1. **Image Pull Errors**
    - Verify image repository and tag
    - Check container registry access

2. **Resource Limits**
    - Monitor resource usage
    - Adjust limits based on actual needs

3. **Route Issues**
    - Verify route hostname
    - Check TLS configuration