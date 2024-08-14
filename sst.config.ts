import { APIStack } from "./stacks/APIStack";
import { DynamoDBStack } from "./stacks/DynamoDBStack";
import { OpenSearchServerlessStack } from "./stacks/OpenSearchServerlessStack";
import { VPCStack } from "./stacks/VPCStack";
import type { SSTConfig } from "sst";

export default {
  config(_input) {
    return {
      name: "real-world-golang-sst-lambda-dynamodb",
      region: "eu-west-1"
    };
  },

  stacks(app) {
    app.stack(VPCStack);
    app.stack(DynamoDBStack);
    app.stack(APIStack);
    app.stack(OpenSearchServerlessStack);
  }
} satisfies SSTConfig;
