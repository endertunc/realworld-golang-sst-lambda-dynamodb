import * as ec2 from "aws-cdk-lib/aws-ec2";
import type { StackContext } from "sst/constructs";

export function VPCStack({ stack }: StackContext) {
  const vpc = new ec2.Vpc(stack, "real-world-vpc", {
    ipAddresses: ec2.IpAddresses.cidr("10.16.0.0/16"),
    maxAzs: 3,
    natGateways: 1,
    subnetConfiguration: [
      {
        cidrMask: 20,
        name: "real-world-public-subnet",
        subnetType: ec2.SubnetType.PUBLIC
      },
      {
        cidrMask: 20,
        name: "real-world-private-subnet",
        subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS
      }
    ]
  });

  vpc.addInterfaceEndpoint("SecretManagerEndpoint", {
    service: ec2.InterfaceVpcEndpointAwsService.SECRETS_MANAGER
  });

  vpc.addGatewayEndpoint("DynamodDBEndpoint", {
    service: ec2.GatewayVpcEndpointAwsService.DYNAMODB
  });

  return {
    vpc: vpc,
    subnetIds: vpc.privateSubnets,
    securityGroupId: vpc.vpcDefaultSecurityGroup
  };
}
