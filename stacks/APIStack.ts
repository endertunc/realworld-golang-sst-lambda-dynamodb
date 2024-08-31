import * as ec2 from "aws-cdk-lib/aws-ec2";
import { Effect, ManagedPolicy, PolicyStatement } from "aws-cdk-lib/aws-iam";
import { IManagedPolicy } from "aws-cdk-lib/aws-iam/lib/managed-policy";
import { Api, Function, use } from "sst/constructs";
import { OpenSearchStack } from "./OpenSearchStack";
import { VPCStack } from "./VPCStack";
import type { StackContext } from "sst/constructs";

export function APIStack({ stack }: StackContext) {
  const { vpc, privateSubnets, securityGroupId } = use(VPCStack);
  // const { opensearchBackendRole } = use(OpenSearchStack);

  const lambdaSecurityGroupId = ec2.SecurityGroup.fromSecurityGroupId(
    stack,
    "real-world-lambda-security-group-id",
    securityGroupId
  );

  // opensearchBackendRole.addManagedPolicy(
  //   ManagedPolicy.fromAwsManagedPolicyName("service-role/AWSLambdaBasicExecutionRole")
  // );
  // opensearchBackendRole.addManagedPolicy(
  //   ManagedPolicy.fromAwsManagedPolicyName("service-role/AWSLambdaVPCAccessExecutionRole")
  // );

  function createLambdaFunction(functionName: string, handler: string) {
    return new Function(stack, functionName, {
      // role: opensearchBackendRole,
      runtime: "go",
      handler: `cmd/functions/${handler}`,
      vpc: vpc,
      logRetention: "one_week",
      vpcSubnets: {
        subnets: privateSubnets
      },
      securityGroups: [lambdaSecurityGroupId],
      environment: {
        // ...(process.env.AWS_PROFILE && { AWS_PROFILE: process.env.AWS_PROFILE }),
        SECRET_NAME: "private-key",
        OPENSEARCH_URL:
          "https://search-realworldopense-0hqrnggvst7d-ykwqy2yzwcqsiz2jvbrs7g5yya.eu-west-1.es.amazonaws.com"
      }
    });
  }

  const helloWorld = createLambdaFunction("hello-world", "hello_world/hello_world.go");
  // const loginUser = createLambdaFunction("login-user", "login_user/login_user.go");
  // const registerUser = createLambdaFunction("register-user", "register_user/register_user.go");
  // const updateUser = createLambdaFunction("update-user", "update_user/update_user.go");
  // const getCurrentUser = createLambdaFunction("get-current-user", "get_current_user/get_current_user.go");
  // const getUserProfile = createLambdaFunction("get-user-profile", "get_user_profile/get_user_profile.go");
  //
  // const followUser = createLambdaFunction("follow-user", "follow_user/follow_user.go");
  // const unfollowUser = createLambdaFunction("unfollow-user", "unfollow_user/unfollow_user.go");
  //
  // const addArticle = createLambdaFunction("add-article", "add_article/add_article.go");
  // const updateArticle = createLambdaFunction("update-article", "update_article/update_article.go");
  // const getArticle = createLambdaFunction("get-article", "get_article/get_article.go");
  // const deleteArticle = createLambdaFunction("delete-article", "delete_article/delete_article.go");
  //
  // const favoriteArticle = createLambdaFunction("favorite-article", "favorite_article/favorite_article.go");
  // const unfavoriteArticle = createLambdaFunction("unfavorite-article", "unfavorite_article/unfavorite_article.go");
  //
  // const addComment = createLambdaFunction("add-comment", "add_comment/add_comment.go");
  // const deleteComment = createLambdaFunction("delete-comment", "delete_comment/delete_comment.go");
  // const getArticleComments = createLambdaFunction(
  //   "get-article-comments",
  //   "get_article_comments/get_article_comments.go"
  // );

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
    // prettier-ignore
    routes: {
      "GET    /api/hello-world":                  helloWorld,
      // "POST   /api/users/login":                  loginUser,
      // "POST   /api/users":                        registerUser,
      // "GET    /api/user":                         getCurrentUser,
      // "PUT    /api/user":                         updateUser,
      // "GET    /api/profiles/:username":           getUserProfile,
      // "POST   /api/profiles/:username/follow":    followUser,
      // "DELETE /api/profiles/:username/follow":    unfollowUser,
      // "POST   /api/articles":                     addArticle,
      // "PUT    /api/articles/:slug":               updateArticle,
      // "GET    /api/articles/:slug":               getArticle,
      // "DELETE /api/articles/:slug":               deleteArticle,
      // "POST   /api/articles/:slug/favorite":      favoriteArticle,
      // "DELETE /api/articles/:slug/favorite":      unfavoriteArticle,
      // "POST   /api/articles/:slug/comments":      addComment,
      // "DELETE /api/articles/:slug/comments/:id":  deleteComment,
      // "GET    /api/articles/:slug/comments":      getArticleComments
    }
  });

  const dynamoPolicy = new PolicyStatement({
    actions: ["dynamodb:*"], // ToDo @ender this should be more restrictive
    resources: ["arn:aws:dynamodb:*:*:table/*"] // ToDo @ender this should be more restrictive but this is to make deployment faster
  });

  // Grant the Lambda function access to the OpenSearch domain
  // const openSearchPolicy = new PolicyStatement({
  //   actions: ["es:*"],
  //   resources: [openSearchDomainArn]
  // });

  // Grant the Lambda function access to all OpenSearch domains in the account
  const openSearchPolicy = new PolicyStatement({
    actions: ["es:*"], // ToDo @ender this should be more restrictive
    resources: ["arn:aws:es:*:*:domain/*"] // ToDo @ender this should be more restrictive but this is to make deployment faster
  });

  helloWorld.addToRolePolicy(openSearchPolicy);
  helloWorld.addToRolePolicy(dynamoPolicy);

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

  helloWorld.addToRolePolicy(ossAPIPolicy);

  // const ossDashboardPolicy = new PolicyStatement({
  //   effect: Effect.ALLOW,
  //   actions: ["aoss:DashboardsAccessAll"],
  //   resources: [`arn:aws:aoss:${stack.region}:${stack.account}:dashboards/default`]
  // });
  // userLogin.addToRolePolicy(ossAPIPolicy);

  return {
    apiUrl: realWorldApi.url,
    roleArn: helloWorld.role?.roleArn
  };
}
