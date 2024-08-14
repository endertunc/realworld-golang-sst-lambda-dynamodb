// import { RemovalPolicy } from "aws-cdk-lib";
// import * as ec2 from "aws-cdk-lib/aws-ec2";
// import * as opensearch from "aws-cdk-lib/aws-opensearchservice";
// import type { StackContext } from "sst/constructs";
// import { use } from "sst/constructs";
// import { VPCStack } from "./VPCStack";
//
// export function OpenSearchStack({ stack }: StackContext) {
//   const { vpc, subnetIds, securityGroupId } = use(VPCStack);
//
//   const openSearchDomain = new opensearch.Domain(stack, "real-world-opensearch-domain", {
//     version: opensearch.EngineVersion.OPENSEARCH_2_11,
//     removalPolicy: RemovalPolicy.DESTROY,
//     capacity: {
//       masterNodes: 3,
//       masterNodeInstanceType: "t3.small.search",
//       dataNodes: 3,
//       dataNodeInstanceType: "t3.small.search"
//     },
//     vpc: vpc,
//     zoneAwareness: {
//       enabled: true,
//       availabilityZoneCount: 3
//     },
//     vpcSubnets: [
//       {
//         subnets: subnetIds
//       }
//     ],
//     securityGroups: [
//       ec2.SecurityGroup.fromSecurityGroupId(stack, "real-world-opensearch-imported-security-group-id", securityGroupId)
//     ]
//   });
//
//   return {
//     openSearchEndpoint: openSearchDomain.domainEndpoint
//   };
// }
