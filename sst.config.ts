import { APIStack } from "./stacks/APIStack";
import { DynamoDBStack } from "./stacks/DynamoDBStack";
import { OpenSearchStack } from "./stacks/OpenSearchStack";
import { VPCStack } from "./stacks/VPCStack";
import type { SSTConfig } from "sst";

export default {
  config(_input) {
    return {
      name: "real-world-golang-sst-lambda-dynamodb",
      region: "eu-west-1"
    };
  },

  async stacks(app) {
    app.stack(VPCStack);
    app.stack(DynamoDBStack);
    await app.stack(OpenSearchStack);
    app.stack(APIStack);
  }
} satisfies SSTConfig;
