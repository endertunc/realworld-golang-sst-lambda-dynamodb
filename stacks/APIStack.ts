import * as ec2 from "aws-cdk-lib/aws-ec2";
import { PolicyStatement } from "aws-cdk-lib/aws-iam";
import { FilterCriteria, FilterRule, StartingPosition } from "aws-cdk-lib/aws-lambda";
import { DynamoEventSource } from "aws-cdk-lib/aws-lambda-event-sources";
import { Api, Function, use } from "sst/constructs";
import { DynamoDBStack } from "./DynamoDBStack";
import { OpenSearchStack } from "./OpenSearchStack";
import { VPCStack } from "./VPCStack";
import type { StackContext } from "sst/constructs";

export function APIStack({ stack }: StackContext) {
  const { vpc, privateSubnets, securityGroupId } = use(VPCStack);
  const { openSearchDomain } = use(OpenSearchStack);
  const dynamodbStack = use(DynamoDBStack);

  const lambdaSecurityGroupId = ec2.SecurityGroup.fromSecurityGroupId(
    stack,
    "real-world-lambda-security-group-id",
    securityGroupId
  );

  function createLambdaFunction(functionName: string, handler: string) {
    const lambda = new Function(stack, functionName, {
      runtime: "go",
      handler: `cmd/functions/${handler}`,
      vpc: vpc,
      logRetention: "one_week",
      vpcSubnets: {
        subnets: privateSubnets
      },
      securityGroups: [lambdaSecurityGroupId],
      environment: {
        OPENSEARCH_URL: `https://${openSearchDomain.domainEndpoint}`
      }
    });
    return lambda;
  }

  const dynamoPolicy = new PolicyStatement({
    actions: ["dynamodb:*"], // ToDo @ender this should be more restrictive
    resources: ["arn:aws:dynamodb:*:*:table/*"] // ToDo @ender this should be more restrictive but this is to make deployment faster
  });

  // Grant the Lambda function access to all OpenSearch domains in the account
  const openSearchPolicy = new PolicyStatement({
    // actions: ["es:*"],
    // resources: ["arn:aws:es:*:*:domain/*"]
    actions: ["es:ESHttpGet", "es:ESHttpPost"],
    resources: [`${openSearchDomain.domainArn}/*`] // ToDo @ender "/*" could be "/article"
  });

  const helloWorld = createLambdaFunction("hello-world", "hello_world/hello_world.go");
  [openSearchPolicy, dynamoPolicy].forEach((policy) => helloWorld.addToRolePolicy(policy));

  const loginUser = createLambdaFunction("login-user", "login_user/login_user.go");
  dynamodbStack.userTable.grantReadData(loginUser);

  const registerUser = createLambdaFunction("register-user", "register_user/register_user.go");
  dynamodbStack.userTable.grantWriteData(registerUser);

  const getCurrentUser = createLambdaFunction("get-current-user", "get_current_user/get_current_user.go");
  dynamodbStack.userTable.grantReadData(getCurrentUser);

  const updateUser = createLambdaFunction("update-user", "update_user/update_user.go");
  dynamodbStack.userTable.grantReadWriteData(updateUser);

  const getUserProfile = createLambdaFunction("get-user-profile", "get_user_profile/get_user_profile.go");
  dynamodbStack.userTable.grantReadData(getUserProfile);
  dynamodbStack.followerTable.grantReadData(getUserProfile);

  const followUser = createLambdaFunction("follow-user", "follow_user/follow_user.go");
  dynamodbStack.userTable.grantReadData(followUser);
  dynamodbStack.followerTable.grantWriteData(followUser);

  const unfollowUser = createLambdaFunction("unfollow-user", "unfollow_user/unfollow_user.go");
  dynamodbStack.userTable.grantReadData(unfollowUser);
  dynamodbStack.followerTable.grantWriteData(unfollowUser);

  const postArticle = createLambdaFunction("post-article", "post_article/post_article.go");
  dynamodbStack.articleTable.grantWriteData(postArticle);
  dynamodbStack.userTable.grantReadData(postArticle);

  const updateArticle = createLambdaFunction("update-article", "update_article/update_article.go");
  dynamodbStack.articleTable.grantReadWriteData(updateArticle);
  dynamodbStack.userTable.grantReadData(updateArticle);
  dynamodbStack.favoritedTable.grantReadData(updateArticle);

  const getArticle = createLambdaFunction("get-article", "get_article/get_article.go");
  dynamodbStack.articleTable.grantReadData(getArticle);
  dynamodbStack.userTable.grantReadData(getArticle);
  dynamodbStack.followerTable.grantReadData(getArticle);
  dynamodbStack.favoritedTable.grantReadData(getArticle);

  const getUserFeed = createLambdaFunction("get-user-feed", "get_user_feed/get_user_feed.go");
  dynamodbStack.feedTable.grantReadData(getUserFeed);
  dynamodbStack.userTable.grantReadData(getUserFeed);
  dynamodbStack.articleTable.grantReadData(getUserFeed);
  dynamodbStack.followerTable.grantReadData(getUserFeed);
  dynamodbStack.favoritedTable.grantReadData(getUserFeed);

  const listArticles = createLambdaFunction("list-articles", "list_articles/list_articles.go");
  dynamodbStack.articleTable.grantReadData(listArticles);
  dynamodbStack.userTable.grantReadData(listArticles);
  dynamodbStack.favoritedTable.grantReadData(listArticles);
  dynamodbStack.followerTable.grantReadData(listArticles);
  listArticles.addToRolePolicy(openSearchPolicy);

  const deleteArticle = createLambdaFunction("delete-article", "delete_article/delete_article.go");
  dynamodbStack.articleTable.grantReadWriteData(deleteArticle);

  const favoriteArticle = createLambdaFunction("favorite-article", "favorite_article/favorite_article.go");
  dynamodbStack.favoritedTable.grantWriteData(favoriteArticle);
  dynamodbStack.articleTable.grantReadWriteData(favoriteArticle);
  dynamodbStack.userTable.grantReadData(favoriteArticle);
  dynamodbStack.followerTable.grantReadData(favoriteArticle);

  const unfavoriteArticle = createLambdaFunction("unfavorite-article", "unfavorite_article/unfavorite_article.go");
  dynamodbStack.favoritedTable.grantWriteData(unfavoriteArticle);
  dynamodbStack.articleTable.grantReadWriteData(unfavoriteArticle);
  dynamodbStack.userTable.grantReadData(unfavoriteArticle);
  dynamodbStack.followerTable.grantReadData(unfavoriteArticle);

  const addComment = createLambdaFunction("add-comment", "add_comment/add_comment.go");
  dynamodbStack.commentTable.grantWriteData(addComment);
  dynamodbStack.articleTable.grantReadData(addComment);
  dynamodbStack.userTable.grantReadData(addComment);

  const deleteComment = createLambdaFunction("delete-comment", "delete_comment/delete_comment.go");
  dynamodbStack.commentTable.grantWriteData(deleteComment);

  const getArticleComments = createLambdaFunction(
    "get-article-comments",
    "get_article_comments/get_article_comments.go"
  );
  dynamodbStack.commentTable.grantReadData(getArticleComments);

  const getTags = createLambdaFunction("get-tags", "get_tags/get_tags.go");
  getTags.addToRolePolicy(openSearchPolicy);

  const swagger = createLambdaFunction("swagger-ui", "swagger/swagger_ui.go");

  const realWorldApi = new Api(stack, "real-world-api", {
    // prettier-ignore
    routes: {
      "GET    /api/hello-world":                    helloWorld,
      "POST   /api/users/login":                    loginUser,
      "POST   /api/users":                          registerUser,
      "GET    /api/user":                           getCurrentUser,
      "PUT    /api/user":                           updateUser,
      "GET    /api/profiles/{username}":            getUserProfile,
      "POST   /api/profiles/{username}/follow":     followUser,
      "DELETE /api/profiles/{username}/follow":     unfollowUser,
      "POST   /api/articles":                       postArticle,
      "PUT    /api/articles/{slug}":                updateArticle,
      "GET    /api/articles":                       listArticles,
      "GET    /api/articles/feed":                  getUserFeed,
      "GET    /api/articles/{slug}":                getArticle,
      "DELETE /api/articles/{slug}":                deleteArticle,
      "POST   /api/articles/{slug}/favorite":       favoriteArticle,
      "DELETE /api/articles/{slug}/favorite":       unfavoriteArticle,
      "POST   /api/articles/{slug}/comments":       addComment,
      "DELETE /api/articles/{slug}/comments/{id}":  deleteComment,
      "GET    /api/articles/{slug}/comments":       getArticleComments,
      "GET    /api/tags":                           getTags,
      "GET    /docs":                               swagger,
      "GET    /docs/spec.json":                     swagger,
    }
  });

  const userFeedEventHandler = createLambdaFunction("feed-event-handler", "user_feed/event_handler.go");
  dynamodbStack.feedTable.grantWriteData(userFeedEventHandler);
  dynamodbStack.followerTable.grantReadData(userFeedEventHandler);
  dynamodbStack.articleTable.grantStreamRead(userFeedEventHandler);

  userFeedEventHandler.addEventSource(
    new DynamoEventSource(dynamodbStack.articleTable, {
      enabled: true,
      startingPosition: StartingPosition.LATEST,
      filters: [
        FilterCriteria.filter({
          eventName: FilterRule.isEqual("INSERT"),
          dynamodb: {
            Keys: {
              pk: { S: [{ "anything-but": { prefix: "slug#" } }] }
            }
          }
        })
      ],
      reportBatchItemFailures: true,
      retryAttempts: 5,
      onFailure: undefined // ToDo @ender add DeadLetterQueue
    })
  );

  stack.addOutputs({
    API_URL: realWorldApi.url
  });

  return {
    apiUrl: realWorldApi.url,
    roleArn: helloWorld.role?.roleArn
  };
}
