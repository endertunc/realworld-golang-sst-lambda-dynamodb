import * as ec2 from "aws-cdk-lib/aws-ec2";
import { LookupMachineImage, Peer, Port } from "aws-cdk-lib/aws-ec2";
// import * as iam from "aws-cdk-lib/aws-iam";
import { FckNatInstanceProvider } from "cdk-fck-nat";
import { getPrefixedResourceName } from "./helpers";
import type { StackContext } from "sst/constructs";

export function VPCStack({ stack, app }: StackContext) {
  const natGatewayProvider = new FckNatInstanceProvider({
    instanceType: new ec2.InstanceType("t3.micro"),
    machineImage: new LookupMachineImage({
      name: "fck-nat-al2023-*-x86_64-ebs",
      owners: ["568608671756"]
    })
  });

  const vpc = new ec2.Vpc(stack, getPrefixedResourceName(app, "vpc"), {
    ipAddresses: ec2.IpAddresses.cidr("10.16.0.0/16"),
    maxAzs: 3, // use single az to reduce cost during development
    natGateways: 1,
    natGatewayProvider: natGatewayProvider,
    subnetConfiguration: [
      {
        cidrMask: 20,
        name: getPrefixedResourceName(app, "public-subnet"),
        subnetType: ec2.SubnetType.PUBLIC
      },
      {
        cidrMask: 20,
        name: getPrefixedResourceName(app, "private-subnet"),
        subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS
      }
    ]
  });
  natGatewayProvider.securityGroup.addIngressRule(Peer.ipv4(vpc.vpcCidrBlock), Port.allTraffic());

  vpc.addGatewayEndpoint(getPrefixedResourceName(app, "dynamodb-gateway-endpoint"), {
    service: ec2.GatewayVpcEndpointAwsService.DYNAMODB
  });

  return {
    vpc: vpc,
    privateSubnets: vpc.selectSubnets({ subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS, onePerAz: true }).subnets,
    publicSubnets: vpc.selectSubnets({ subnetType: ec2.SubnetType.PUBLIC, onePerAz: true }).subnets,
    securityGroupId: vpc.vpcDefaultSecurityGroup
  };
}
