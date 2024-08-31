import * as cdk from "aws-cdk-lib";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import type { StackContext } from "sst/constructs";

export function DynamoDBStack({ stack }: StackContext) {
  const commonTableProps = {
    billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
    removalPolicy: cdk.RemovalPolicy.DESTROY
  };

  const userTable = new dynamodb.Table(stack, "user", {
    ...commonTableProps,
    tableName: "user",
    partitionKey: {
      name: "pk",
      type: dynamodb.AttributeType.STRING
    },
    sortKey: {
      name: "sk",
      type: dynamodb.AttributeType.STRING
    }
  });

  userTable.addGlobalSecondaryIndex({
    indexName: "user_email_gsi",
    projectionType: dynamodb.ProjectionType.ALL,
    partitionKey: {
      name: "email",
      type: dynamodb.AttributeType.STRING
    }
  });

  const articleTable = new dynamodb.Table(stack, "article", {
    ...commonTableProps,
    tableName: "article",
    partitionKey: {
      name: "pk",
      type: dynamodb.AttributeType.STRING
    },
    pointInTimeRecovery: true,
    stream: dynamodb.StreamViewType.NEW_AND_OLD_IMAGES
  });

  articleTable.addGlobalSecondaryIndex({
    indexName: "article_slug_gsi",
    projectionType: dynamodb.ProjectionType.ALL,
    partitionKey: {
      name: "slug",
      type: dynamodb.AttributeType.STRING
    }
  });

  const feedTable = new dynamodb.Table(stack, "feed", {
    ...commonTableProps,
    tableName: "feed",
    partitionKey: {
      name: "userId",
      type: dynamodb.AttributeType.STRING
    },
    sortKey: {
      name: "articleId",
      type: dynamodb.AttributeType.STRING
    }
  });

  const commentTable = new dynamodb.Table(stack, "comment", {
    ...commonTableProps,
    tableName: "comment",
    partitionKey: {
      name: "commentId",
      type: dynamodb.AttributeType.STRING
    },
    sortKey: {
      name: "articleId",
      type: dynamodb.AttributeType.STRING
    }
  });

  commentTable.addGlobalSecondaryIndex({
    indexName: "comment_article_gsi",
    projectionType: dynamodb.ProjectionType.ALL,
    partitionKey: {
      name: "articleId",
      type: dynamodb.AttributeType.STRING
    }
  });

  const favoritetTable = new dynamodb.Table(stack, "favorite", {
    ...commonTableProps,
    tableName: "favorite",
    partitionKey: {
      name: "userId",
      type: dynamodb.AttributeType.STRING
    },
    sortKey: {
      name: "articleId",
      type: dynamodb.AttributeType.STRING
    }
  });

  // ToDo - do we need this???
  favoritetTable.addGlobalSecondaryIndex({
    indexName: "favorite_article_gsi",
    projectionType: dynamodb.ProjectionType.ALL,
    partitionKey: {
      name: "articleId",
      type: dynamodb.AttributeType.STRING
    }
  });

  const followerTable = new dynamodb.Table(stack, "follower", {
    ...commonTableProps,
    tableName: "follower",
    partitionKey: {
      name: "follower",
      type: dynamodb.AttributeType.STRING
    },
    sortKey: {
      name: "followee",
      type: dynamodb.AttributeType.STRING
    }
  });

  return {
    userTableName: userTable.tableName,
    userTableArn: userTable.tableArn
  };
}
