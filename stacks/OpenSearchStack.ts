import * as cognitoIdp from "@aws-cdk/aws-cognito-identitypool-alpha";
// import { fromNodeProviderChain } from "@aws-sdk/credential-providers";
// import { Client } from "@opensearch-project/opensearch/api/new";
// import { AwsSigv4Signer } from "@opensearch-project/opensearch/lib/aws";
import { RemovalPolicy } from "aws-cdk-lib";
import * as cognito from "aws-cdk-lib/aws-cognito";
import * as iam from "aws-cdk-lib/aws-iam";
import * as logs from "aws-cdk-lib/aws-logs";
import { RetentionDays } from "aws-cdk-lib/aws-logs";
import * as opensearch from "aws-cdk-lib/aws-opensearchservice";
import * as osis from "aws-cdk-lib/aws-osis";
import { getPrefixedResourceName } from "./helpers";
import type { StackContext } from "sst/constructs";

export async function OpenSearchStack({ stack, app }: StackContext) {
  const userPool = new cognito.UserPool(stack, getPrefixedResourceName(app, "open-search-dashboard-user-pool"), {
    userPoolName: getPrefixedResourceName(app, "open-search-dashboard-user-pool"),
    selfSignUpEnabled: false,
    signInAliases: { username: true, email: false },
    autoVerify: { email: true },
    mfa: cognito.Mfa.OFF,
    passwordPolicy: {
      minLength: 6,
      requireLowercase: false,
      requireUppercase: false,
      requireDigits: false,
      requireSymbols: false
    },
    removalPolicy: RemovalPolicy.DESTROY
  });

  new cognito.UserPoolDomain(stack, getPrefixedResourceName(app, "opensearch-user-pool-domain"), {
    userPool,
    cognitoDomain: {
      domainPrefix: getPrefixedResourceName(app, "opensearch-dashboard")
    }
  });

  const idPool = new cognitoIdp.IdentityPool(stack, getPrefixedResourceName(app, "open-search-dashboard-idp"), {
    identityPoolName: getPrefixedResourceName(app, "open-search-dashboard-idp"),
    allowUnauthenticatedIdentities: true,
    authenticationProviders: {
      userPools: [new cognitoIdp.UserPoolAuthenticationProvider({ userPool })]
    }
  });

  const dashboardRole = new iam.Role(stack, getPrefixedResourceName(app, "opensearch-dashboard-role"), {
    assumedBy: new iam.ServicePrincipal("opensearchservice.amazonaws.com"),
    managedPolicies: [iam.ManagedPolicy.fromAwsManagedPolicyName("AmazonOpenSearchServiceCognitoAccess")]
  });

  const openSearchDomain = new opensearch.Domain(stack, getPrefixedResourceName(app, "opensearch-domain"), {
    version: opensearch.EngineVersion.OPENSEARCH_2_11,
    removalPolicy: RemovalPolicy.DESTROY,
    capacity: {
      masterNodes: 2,
      masterNodeInstanceType: "t3.small.search",
      dataNodes: 1,
      dataNodeInstanceType: "t3.small.search",
      warmNodes: 0
    },
    zoneAwareness: {
      enabled: false
    },
    encryptionAtRest: {
      enabled: true
    },
    nodeToNodeEncryption: true,
    enforceHttps: true,
    cognitoDashboardsAuth: {
      userPoolId: userPool.userPoolId,
      identityPoolId: idPool.identityPoolId,
      role: dashboardRole
    }
  });

  const ingestionPipelineRole = new iam.Role(stack, getPrefixedResourceName(app, "ingestion-pipeline-role"), {
    assumedBy: new iam.ServicePrincipal("osis-pipelines.amazonaws.com", {
      // https://docs.aws.amazon.com/opensearch-service/latest/developerguide/pipeline-domain-access.html
      conditions: {
        StringEquals: {
          "aws:SourceAccount": stack.account
        },
        ArnLike: {
          "aws:SourceArn": `arn:aws:osis:${stack.region}:${stack.account}:pipeline/*`
        }
      }
    }),
    inlinePolicies: {
      ingestionPipeline: new iam.PolicyDocument({
        statements: [
          new iam.PolicyStatement({
            actions: ["es:DescribeDomain"],
            resources: [openSearchDomain.domainArn]
          }),
          new iam.PolicyStatement({
            actions: ["es:ESHttp*"],
            resources: [`${openSearchDomain.domainArn}/*`]
          }),
          new iam.PolicyStatement({
            actions: ["dynamodb:DescribeTable"],
            resources: [`arn:aws:dynamodb:${stack.region}:${stack.account}:table/article`]
          }),
          new iam.PolicyStatement({
            actions: [
              "dynamodb:DescribeTable",
              "dynamodb:DescribeStream",
              "dynamodb:GetRecords",
              "dynamodb:GetShardIterator"
            ],
            resources: [`arn:aws:dynamodb:${stack.region}:${stack.account}:table/article/stream/*`]
          })
        ]
      })
    }
  });

  openSearchDomain.addAccessPolicies(
    new iam.PolicyStatement({
      principals: [ingestionPipelineRole],
      actions: ["es:DescribeDomain", "es:ESHttp*"],
      resources: [`${openSearchDomain.domainArn}/*`]
    })
  );

  const authenticatedRole = new iam.Policy(stack, getPrefixedResourceName(app, "idp-authenticated-policy"), {
    statements: [
      new iam.PolicyStatement({
        effect: iam.Effect.ALLOW,
        resources: [openSearchDomain.domainArn + "/*"],
        actions: ["es:*"]
      })
    ]
  });
  idPool.authenticatedRole.attachInlinePolicy(authenticatedRole);

  const ingestionPipelineLogGroup = new logs.LogGroup(stack, getPrefixedResourceName(app, "osis-article-log-group"), {
    logGroupName: `/aws/vendedlogs/OpenSearchIngestion/article/${getPrefixedResourceName(app)}-audit-logs`,
    removalPolicy: RemovalPolicy.DESTROY,
    retention: RetentionDays.ONE_DAY
  });

  new osis.CfnPipeline(stack, getPrefixedResourceName(app, "osis-article-pipeline"), {
    pipelineName: getPrefixedResourceName(app, "article"),
    minUnits: 1,
    maxUnits: 1,
    logPublishingOptions: {
      isLoggingEnabled: true,
      cloudWatchLogDestination: {
        logGroup: ingestionPipelineLogGroup.logGroupName
      }
    },
    pipelineConfigurationBody: /* yaml */ `
version: "2"
dynamodb-pipeline:
  source:
    dynamodb:
      tables:
        # REQUIRED: Supply the DynamoDB table ARN and whether export or stream processing is needed, or both
        - table_arn: arn:aws:dynamodb:${stack.region}:${stack.account}:table/article
          stream:
            start_position: LATEST # TRIM_HORIZON does not work now. It will make pipeline unstable and cannot be deployed.
      aws:
        # REQUIRED: Provide the role to assume that has the necessary permissions to DynamoDB, OpenSearch, and S3.
        sts_role_arn: ${ingestionPipelineRole.roleArn}
        # Provide the region to use for aws credentials
        region: ${stack.region}
  sink:
    - opensearch:
        # REQUIRED: Provide an AWS OpenSearch endpoint
        hosts:
          - https://${openSearchDomain.domainEndpoint}
        index: article
        index_type: custom
        document_id: \${getMetadata("primary_key")}
        action: \${getMetadata("opensearch_action")}
        document_version: \${getMetadata("document_version")}
        document_version_type: external
        aws:
          # REQUIRED: Provide a Role ARN with access to the domain. This role should have a trust relationship with osis-pipelines.amazonaws.com
          sts_role_arn: ${ingestionPipelineRole.roleArn}
          # Provide the region of the domain.
          region: ${stack.region}
      `
  });

  stack.addOutputs({
    OPENSEARCH_URL: `https://${openSearchDomain.domainEndpoint}`
  });

  return {
    openSearchDomain
  };
}
