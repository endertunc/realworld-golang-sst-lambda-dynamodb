import * as iam from "aws-cdk-lib/aws-iam";
import * as opensearchserverless from "aws-cdk-lib/aws-opensearchserverless";
import { use } from "sst/constructs";
import { APIStack } from "./APIStack";
import type { StackContext } from "sst/constructs";

export function OpenSearchServerlessStack({ stack }: StackContext) {
  const { roleArn } = use(APIStack);
  const oss = new opensearchserverless.CfnCollection(stack, "real-world-article-collection", {
    name: "article",
    type: "SEARCH"
  });

  const dashboardAccessRole = "AWSReservedSSO_AdministratorAccess_ed5e5518843fe6b0";

  const ossNetworkSecurityPolicy = new opensearchserverless.CfnSecurityPolicy(stack, "rw-oss-network-policy", {
    name: "rw-oss-network-policy",
    type: "network",
    policy: JSON.stringify([
      {
        Rules: [
          { ResourceType: "collection", Resource: ["collection/article"] },
          { ResourceType: "dashboard", Resource: ["collection/article"] }
        ],
        AllowFromPublic: true
      }
    ])
  });

  const ossEncryptionSecurityPolicy = new opensearchserverless.CfnSecurityPolicy(stack, "rw-oss-encryption-policy", {
    name: "rw-oss-encryption-policy",
    type: "encryption",
    policy: JSON.stringify({
      Rules: [
        {
          ResourceType: "collection",
          Resource: ["collection/article"]
        }
      ],
      AWSOwnedKey: true
    })
  });

  const ossDataAccessPolicy = new opensearchserverless.CfnAccessPolicy(stack, "rw-oss-access-policy", {
    name: "rw-oss-access-policy",
    type: "data",
    policy: JSON.stringify([
      {
        Rules: [
          {
            ResourceType: "collection",
            Resource: ["collection/article"],
            Permission: ["aoss:*"] // ToDo @ender this should be more restrictive
          },
          {
            ResourceType: "index",
            Resource: ["index/*/*"],
            Permission: ["aoss:*"] // ToDo @ender this should be more restrictive
          }
        ],
        Principal: [roleArn, `arn:aws:sts::${stack.account}:assumed-role/${dashboardAccessRole}/*`]
      }
    ])
  });

  oss.addDependency(ossNetworkSecurityPolicy);
  oss.addDependency(ossEncryptionSecurityPolicy);
  oss.addDependency(ossDataAccessPolicy);

  return {
    openSearchDashboardEndpoint: oss.attrDashboardEndpoint,
    openSearchCollectionEndpoint: oss.attrCollectionEndpoint
  };
}
