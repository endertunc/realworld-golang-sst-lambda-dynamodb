// import { RemovalPolicy } from "aws-cdk-lib";
// import * as iam from "aws-cdk-lib/aws-iam";
// import * as logs from "aws-cdk-lib/aws-logs";
// import { RetentionDays } from "aws-cdk-lib/aws-logs";
// import * as osis from "aws-cdk-lib/aws-osis";
// import { use } from "sst/constructs";
// import { OpenSearchStack } from "./OpenSearchStack";
// import type { StackContext } from "sst/constructs";
//
// export function OpenSearchIngestionPipeline({ stack }: StackContext) {
//   const { openSearchDomain } = use(OpenSearchStack);
//
//   const ingestionPipelineRole = new iam.Role(stack, "real-world-ingestion-pipeline-role", {
//     assumedBy: new iam.ServicePrincipal("osis-pipelines.amazonaws.com", {
//       // https://docs.aws.amazon.com/opensearch-service/latest/developerguide/pipeline-domain-access.html
//       conditions: {
//         StringEquals: {
//           "aws:SourceAccount": stack.account
//         },
//         ArnLike: {
//           "aws:SourceArn": `arn:aws:osis:${stack.region}:${stack.account}:pipeline/*`
//         }
//       }
//     }),
//     inlinePolicies: {
//       ingestionPipeline: new iam.PolicyDocument({
//         statements: [
//           new iam.PolicyStatement({
//             actions: ["es:DescribeDomain"],
//             resources: [openSearchDomain.domainArn]
//           }),
//           new iam.PolicyStatement({
//             actions: ["es:ESHttp*"],
//             resources: [`${openSearchDomain.domainArn}/*`]
//           }),
//           new iam.PolicyStatement({
//             actions: ["dynamodb:DescribeStream", "dynamodb:GetRecords", "dynamodb:GetShardIterator"],
//             resources: [`arn:aws:dynamodb:${stack.region}:${stack.account}:table/article/stream/*`]
//           })
//         ]
//       })
//     }
//   });
//
//   openSearchDomain.addAccessPolicies(
//     new iam.PolicyStatement({
//       principals: [ingestionPipelineRole],
//       actions: ["es:DescribeDomain", "es:ESHttp*"],
//       resources: [`${openSearchDomain.domainArn}/*`]
//     }),
//     new iam.PolicyStatement({
//       effect: iam.Effect.ALLOW,
//       actions: ["es:ESHttp*"],
//       principals: [new iam.AccountPrincipal(stack.account)],
//       resources: [`${openSearchDomain.domainArn}/*`]
//     })
//   );
//
//   const ingestionPipelineLogGroup = new logs.LogGroup(stack, "real-world-osis-article-log-group", {
//     logGroupName: `/aws/vendedlogs/OpenSearchIngestion/article/audit-logs`,
//     removalPolicy: RemovalPolicy.DESTROY,
//     retention: RetentionDays.ONE_DAY
//   });
//
//   new osis.CfnPipeline(stack, "real-world-osis-article-pipeline", {
//     pipelineName: "osis-article-pipeline",
//     minUnits: 1,
//     maxUnits: 1,
//     logPublishingOptions: {
//       isLoggingEnabled: true,
//       cloudWatchLogDestination: {
//         logGroup: ingestionPipelineLogGroup.logGroupName
//       }
//     },
//     pipelineConfigurationBody: /* yaml */ `
// version: "2"
// dynamodb-pipeline:
//   source:
//     dynamodb:
//       tables:
//         # REQUIRED: Supply the DynamoDB table ARN and whether export or stream processing is needed, or both
//         - table_arn: arn:aws:dynamodb:${stack.region}:${stack.account}:table/article
//           # Remove the stream block if only export is needed
//           stream:
//             start_position: LATEST # TRIM_HORIZON does not work now. It will make pipeline unstable and cannot be deployed.
//       aws:
//         # REQUIRED: Provide the role to assume that has the necessary permissions to DynamoDB, OpenSearch, and S3.
//         sts_role_arn: ${ingestionPipelineRole.roleArn}
//         # Provide the region to use for aws credentials
//         region: ${stack.region}
//   sink:
//     - opensearch:
//         # REQUIRED: Provide an AWS OpenSearch endpoint
//         hosts:
//           - https://${openSearchDomain.domainEndpoint}
//         index: article
//         index_type: custom
//         document_id: \${getMetadata("primary_key")}
//         action: \${getMetadata("opensearch_action")}
//         document_version: \${getMetadata("document_version")}
//         document_version_type: external
//         aws:
//           # REQUIRED: Provide a Role ARN with access to the domain. This role should have a trust relationship with osis-pipelines.amazonaws.com
//           sts_role_arn: ${ingestionPipelineRole.roleArn}
//           # Provide the region of the domain.
//           region: ${stack.region}
//       `
//   });
// }
