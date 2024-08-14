// import fs = require("fs");
// import cdk = require("aws-cdk-lib");
// import lambda = require("aws-cdk-lib/aws-lambda");
// import customResources = require("aws-cdk-lib/custom-resources");
// import type { StackContext } from "sst/constructs";
//
// export interface MyCustomResourceProps {
//   message: string;
// }
//
// export function DynamoDBStack({ stack }: StackContext) {
//   const fn = new lambda.SingletonFunction(stack, "Singleton", {
//     uuid: "f7d4f730-4ee1-11e8-9c2d-fa7ae01bbebc",
//     code: new lambda.InlineCode(fs.readFileSync("custom-resource-handler.py", { encoding: "utf-8" })),
//     handler: "index.main",
//     timeout: cdk.Duration.seconds(300),
//     runtime: lambda.Runtime.PYTHON_3_9
//   });
//
//   const provider = new customResources.Provider(stack, "Provider", {
//     onEventHandler: fn
//   });
//
//   const resource = new cdk.CustomResource(stack, "Resource", {
//     serviceToken: provider.serviceToken
//   });
// }
