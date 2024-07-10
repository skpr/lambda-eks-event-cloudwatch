Lambda: EKS Event CloudWatch
----------------------------

A Lambda for forwarding CloudWatch Alarms to EKS Events.

```mermaid
flowchart LR
   CloudWatch --> Lambda
   Lambda --> EKS
   EKS --> Kubectl
```

## Requirements

### AWS IAM Permissions

```
TBD
```

### CloudWatch Alarm Tags

```
TBD
```
