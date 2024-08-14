import * as ec2 from "aws-cdk-lib/aws-ec2";
import { Effect, PolicyStatement } from "aws-cdk-lib/aws-iam";
import { Api, Function, use } from "sst/constructs";
import { DynamoDBStack } from "./DynamoDBStack";
import { VPCStack } from "./VPCStack";
import type { StackContext } from "sst/constructs";

export function APIStack({ stack }: StackContext) {
  const { vpc, subnetIds, securityGroupId } = use(VPCStack);
  const { userTableName, userTableArn } = use(DynamoDBStack);

  const userLogin = new Function(stack, "user-login", {
    runtime: "go",
    handler: "cmd/functions/hello_world/hello_world.go",
    vpc: vpc,
    logRetention: "one_week",
    vpcSubnets: {
      subnets: subnetIds
    },
    securityGroups: [
      ec2.SecurityGroup.fromSecurityGroupId(stack, "real-world-user-login-security-group-id", securityGroupId)
    ],
    environment: {
      // ...(process.env.AWS_PROFILE && { AWS_PROFILE: process.env.AWS_PROFILE }),
      SECRET_NAME: "private-key",
      TABLE_NAME: userTableName,
      OPENSEARCH_URL: "https://ka9ehtys9fgwac4oju9g.eu-west-1.aoss.amazonaws.com"
    }
  });
  // [profile real-world]
  // region = eu-west-1
  // source_profile = real-world-login
  //
  // [profile real-world-login]
  // sso_start_url = https://d-93675b57d3.awsapps.com/start
  // sso_region = eu-west-1
  // sso_account_id = 571034679658
  // sso_role_name = PowerUserAccess
  // output = json
  // region = eu-west-1

  // const privateKey = new secretsmanager.Secret(stack, 'private-key', {
  //     secretName: 'private-key',
  //     removalPolicy: RemovalPolicy.DESTROY,
  // });

  const realWorldApi = new Api(stack, "real-world-api", {
    routes: {
      "GET /": userLogin
    }
  });

  const dynamoPolicy = new PolicyStatement({
    actions: ["dynamodb:*"], // ToDo @ender this should be more restrictive
    resources: [userTableArn]
  });
  userLogin.addToRolePolicy(dynamoPolicy);

  // const ossAPIPolicy = new PolicyStatement({
  //   effect: Effect.ALLOW,
  //   // actions: ["aoss:APIAccessAll"], // ToDo @ender this should be more restrictive
  //   actions: ["aoss:*"],
  //   resources: [
  //     `arn:aws:aoss:${stack.region}:${stack.account}:collection/article`,
  //     `arn:aws:aoss:${stack.region}:${stack.account}:index/*/*` // ToDe @ender recently added
  //   ]
  // });

  const ossAPIPolicy = new PolicyStatement({
    effect: Effect.ALLOW,
    actions: ["aoss:*"],
    resources: [
      `arn:aws:aoss:${stack.region}:${stack.account}:collection/*`, // ToDo @ender this should be more restrictive
      `arn:aws:aoss:${stack.region}:${stack.account}:index/*/*` // ToDo @ender this should be more restrictive
    ]
  });

  userLogin.addToRolePolicy(ossAPIPolicy);

  // const ossDashboardPolicy = new PolicyStatement({
  //   effect: Effect.ALLOW,
  //   actions: ["aoss:DashboardsAccessAll"],
  //   resources: [`arn:aws:aoss:${stack.region}:${stack.account}:dashboards/default`]
  // });
  // userLogin.addToRolePolicy(ossAPIPolicy);

  return {
    apiUrl: realWorldApi.url,
    roleArn: userLogin.role?.roleArn
  };
}
