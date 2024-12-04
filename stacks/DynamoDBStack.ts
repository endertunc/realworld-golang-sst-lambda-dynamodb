import * as cdk from "aws-cdk-lib";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import { getPrefixedResourceName } from "./helpers";
import type { StackContext } from "sst/constructs";

export function DynamoDBStack({ stack, app }: StackContext) {
  const commonTableProps = {
    billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
    removalPolicy: cdk.RemovalPolicy.DESTROY
  };

  const userTable = new dynamodb.Table(stack, getPrefixedResourceName(app, "user"), {
    ...commonTableProps,
    tableName: "user",
    partitionKey: {
      name: "pk",
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

  userTable.addGlobalSecondaryIndex({
    indexName: "user_username_gsi",
    projectionType: dynamodb.ProjectionType.ALL,
    partitionKey: {
      name: "username",
      type: dynamodb.AttributeType.STRING
    }
  });

  const articleTable = new dynamodb.Table(stack, getPrefixedResourceName(app, "article"), {
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

  articleTable.addGlobalSecondaryIndex({
    indexName: "article_author_gsi",
    projectionType: dynamodb.ProjectionType.ALL,
    partitionKey: {
      name: "authorId",
      type: dynamodb.AttributeType.STRING
    },
    sortKey: {
      name: "createdAt",
      type: dynamodb.AttributeType.NUMBER
    }
  });

  const feedTable = new dynamodb.Table(stack, getPrefixedResourceName(app, "feed"), {
    ...commonTableProps,
    tableName: "feed",
    partitionKey: {
      name: "userId",
      type: dynamodb.AttributeType.STRING
    },
    sortKey: {
      name: "createdAt",
      type: dynamodb.AttributeType.NUMBER
    }
  });

  const commentTable = new dynamodb.Table(stack, getPrefixedResourceName(app, "comment"), {
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

  const favoritedTable = new dynamodb.Table(stack, getPrefixedResourceName(app, "favorite"), {
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

  favoritedTable.addGlobalSecondaryIndex({
    indexName: "favorite_user_id_created_at_gsi",
    projectionType: dynamodb.ProjectionType.ALL,
    partitionKey: {
      name: "userId",
      type: dynamodb.AttributeType.STRING
    },
    sortKey: {
      name: "createdAt",
      type: dynamodb.AttributeType.NUMBER
    }
  });

  const followerTable = new dynamodb.Table(stack, getPrefixedResourceName(app, "follower"), {
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

  followerTable.addGlobalSecondaryIndex({
    indexName: "follower_followee_gsi",
    projectionType: dynamodb.ProjectionType.ALL,
    partitionKey: {
      name: "followee",
      type: dynamodb.AttributeType.STRING
    }
  });

  return {
    articleTable,
    userTable,
    feedTable,
    commentTable,
    favoritedTable,
    followerTable
  };
}
