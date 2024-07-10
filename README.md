Lambda: EKS Event CloudWatch
----------------------------

A Lambda for forwarding CloudWatch Alarms to EKS Events.

```mermaid
flowchart LR
   CloudWatch --> Lambda
   Lambda --> EKS
   EKS --> Kubectl
```
