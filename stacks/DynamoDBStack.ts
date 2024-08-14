import * as cdk from "aws-cdk-lib";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import type { StackContext } from "sst/constructs";

export function DynamoDBStack({ stack }: StackContext) {
  const userTable = new dynamodb.Table(stack, "user", {
    partitionKey: { name: "pk", type: dynamodb.AttributeType.STRING },
    sortKey: { name: "sk", type: dynamodb.AttributeType.STRING },
    billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
    removalPolicy: cdk.RemovalPolicy.DESTROY
  });

  return {
    userTableName: userTable.tableName,
    userTableArn: userTable.tableArn
  };
}
