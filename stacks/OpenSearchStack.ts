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
import type { StackContext } from "sst/constructs";

export async function OpenSearchStack({ stack }: StackContext) {
  // const { vpc, privateSubnets, publicSubnets, securityGroupId, bastionSecurityGroupId } = use(VPCStack);

  // const bastionSecurityGroup = ec2.SecurityGroup.fromSecurityGroupId(
  //   stack,
  //   "real-world-bastion-sg-lookup",
  //   bastionSecurityGroupId
  // );

  // Define the OpenSearch domain
  // const openSearchSecurityGroup = new ec2.SecurityGroup(stack, "real-world-opensearch-sg", {
  //   securityGroupName: "real-world-opensearch-sg",
  //   vpc,
  //   description: "Allow traffic from bastion host",
  //   allowAllOutbound: true
  // });

  // Allow traffic from the bastion host to the OpenSearch domain
  // openSearchSecurityGroup.addIngressRule(bastionSecurityGroup, ec2.Port.tcp(443), "Allow traffic from bastion host");
  // openSearchSecurityGroup.addIngressRule(bastionSecurityGroup, ec2.Port.tcp(80), "Allow traffic from bastion host");

  const userPool = new cognito.UserPool(stack, "real-world-open-search-dashboard-user-pool", {
    userPoolName: "real-world-open-search-dashboard-user-pool",
    selfSignUpEnabled: false,
    signInAliases: { username: true, email: false },
    autoVerify: { email: true },
    mfa: cognito.Mfa.OFF,
    // mfaSecondFactor: { sms: false, otp: true },
    passwordPolicy: {
      minLength: 6,
      requireLowercase: false,
      requireUppercase: false,
      requireDigits: false,
      requireSymbols: false
    },
    removalPolicy: RemovalPolicy.DESTROY
  });

  // const cfnUserPoolUser = new cognito.CfnUserPoolUser(stack, "real-world-opensearch-dashboard-user", {
  //   userPoolId: "userPoolId",
  //   // the properties below are optional
  //   userAttributes: [
  //     {
  //       name: "password",
  //       value: "dashboard-password"
  //     }
  //   ],
  //   username: "dashboard-user"
  // });

  new cognito.UserPoolDomain(stack, "real-world-opensearch-user-pool-domain", {
    userPool,
    cognitoDomain: {
      domainPrefix: "real-world-opensearch-dashboard"
    }
  });

  const idPool = new cognitoIdp.IdentityPool(stack, "real-world-open-search-dashboard-idp", {
    identityPoolName: "real-world-open-search-dashboard-idp",
    allowUnauthenticatedIdentities: true,
    authenticationProviders: {
      userPools: [new cognitoIdp.UserPoolAuthenticationProvider({ userPool })]
    }
  });

  const dashboardRole = new iam.Role(stack, "real-world-opensearch-dashboard-role", {
    assumedBy: new iam.ServicePrincipal("opensearchservice.amazonaws.com"),
    managedPolicies: [iam.ManagedPolicy.fromAwsManagedPolicyName("AmazonOpenSearchServiceCognitoAccess")]
  });

  const openSearchDomain = new opensearch.Domain(stack, "real-world-opensearch-domain", {
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
    // fineGrainedAccessControl: {
    //   masterUserName: "masteruser",
    //   masterUserPassword: SecretValue.unsafePlainText("123qweASD!")
    // },
    // useUnsignedBasicAuth: true,
    // vpc: vpc,
    // vpcSubnets: [
    //   {
    //     subnets: [privateSubnets[0]]
    //   }
    // ],
    // securityGroups: [openSearchSecurityGroup]
  });

  const ingestionPipelineRole = new iam.Role(stack, "real-world-ingestion-pipeline-role", {
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
    // new iam.PolicyStatement({
    //   effect: iam.Effect.ALLOW,
    //   // actions: ["es:*"],
    //   actions: ["es:ESHttp*"],
    //   principals: [idPool.authenticatedRole],
    //   resources: [`${openSearchDomain.domainArn}/*`]
    // })
    // new iam.PolicyStatement({
    //   effect: iam.Effect.ALLOW,
    //   // actions: ["es:*"],
    //   actions: ["es:ESHttp*"],
    //   principals: [new iam.AccountPrincipal(stack.account)],
    //   // principals: [new iam.AnyPrincipal()],
    //   resources: [`${openSearchDomain.domainArn}/*`]
    // })
  );

  const authenticatedRole = new iam.Policy(stack, "real-world-idp-authenticated-policy", {
    statements: [
      new iam.PolicyStatement({
        effect: iam.Effect.ALLOW,
        resources: [openSearchDomain.domainArn + "/*"],
        actions: ["es:*"]
      })
    ]
  });
  idPool.authenticatedRole.attachInlinePolicy(authenticatedRole);

  const ingestionPipelineLogGroup = new logs.LogGroup(stack, "real-world-osis-article-log-group", {
    logGroupName: `/aws/vendedlogs/OpenSearchIngestion/article/audit-logs`,
    removalPolicy: RemovalPolicy.DESTROY,
    retention: RetentionDays.ONE_DAY
  });

  new osis.CfnPipeline(stack, "real-world-osis-article-pipeline", {
    pipelineName: "osis-article-pipeline",
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

  // const opensearchBackendRole = new iam.Role(stack, "real-world-open-search-backend-role", {
  //   assumedBy: new iam.ServicePrincipal("lambda.amazonaws.com")
  // });
  //
  // const client = new Client({
  //   ...AwsSigv4Signer({
  //     region: stack.region,
  //     service: "es",
  //     getCredentials: () => {
  //       const credentialProvider = fromNodeProviderChain();
  //       return credentialProvider();
  //     }
  //   }),
  //   node: openSearchDomain.domainEndpoint // OpenSearch domain URL
  // });
  //
  // const patchRoleResult = await client.security.patchRoleMapping({
  //   role: "all_access",
  //   body: { backend_roles: opensearchBackendRole.roleArn }
  // });
  //
  // console.log("Patch Role Mapping Result: ", patchRoleResult);

  return {
    openSearchDomain
  };
}
