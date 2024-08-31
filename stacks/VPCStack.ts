import * as ec2 from "aws-cdk-lib/aws-ec2";
import { LookupMachineImage, Peer, Port } from "aws-cdk-lib/aws-ec2";
// import * as iam from "aws-cdk-lib/aws-iam";
import { FckNatInstanceProvider } from "cdk-fck-nat";
import type { StackContext } from "sst/constructs";

export function VPCStack({ stack }: StackContext) {
  const natGatewayProvider = new FckNatInstanceProvider({
    instanceType: new ec2.InstanceType("t3.micro"),
    machineImage: new LookupMachineImage({
      name: "fck-nat-al2023-*-x86_64-ebs",
      owners: ["568608671756"]
    })
  });

  // const natGatewayProvider = NatInstanceProviderV2.instanceV2({
  //   instanceType: InstanceType.of(InstanceClass.T3, InstanceSize.MICRO),
  //   machineImage: new LookupMachineImage({
  //     name: "fck-nat-al2023-*-arm64-ebs",
  //     owners: ["568608671756"]
  //   })
  // });

  const vpc = new ec2.Vpc(stack, "real-world-vpc", {
    ipAddresses: ec2.IpAddresses.cidr("10.16.0.0/16"),
    maxAzs: 3, // use single az to reduce cost during development
    natGateways: 1,
    natGatewayProvider: natGatewayProvider,
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

  natGatewayProvider.securityGroup.addIngressRule(Peer.ipv4(vpc.vpcCidrBlock), Port.allTraffic());

  // vpc.addInterfaceEndpoint("SecretManagerEndpoint", {
  //   service: ec2.InterfaceVpcEndpointAwsService.SECRETS_MANAGER
  // });

  vpc.addGatewayEndpoint("DynamodDBEndpoint", {
    service: ec2.GatewayVpcEndpointAwsService.DYNAMODB
  });

  // const publicKeyMaterial = readFileSync("/Users/E.Tunc/.ssh/id_ed25519_bastion_host.pub", "utf-8");
  // const keyPair = new ec2.CfnKeyPair(stack, "bastion-key-pair", {
  //   keyName: "bastion-key-pair",
  //   // keyType: "ed25519",
  //   publicKeyMaterial: publicKeyMaterial,
  //   keyFormat: "pem"
  // });

  // const bastionSecurityGroup = new ec2.SecurityGroup(stack, "BastionSecurityGroup", {
  //   vpc,
  //   description: "Allow SSH access to bastion host",
  //   allowAllOutbound: true
  // });
  // bastionSecurityGroup.addIngressRule(ec2.Peer.anyIpv4(), ec2.Port.tcp(22), "Allow SSH access");

  // Create IAM Role for Bastion Host
  // const bastionRole = new iam.Role(stack, "BastionRole", {
  //   assumedBy: new iam.ServicePrincipal("ec2.amazonaws.com"),
  //   managedPolicies: [
  //     iam.ManagedPolicy.fromAwsManagedPolicyName("AmazonEC2ReadOnlyAccess"),
  //     iam.ManagedPolicy.fromAwsManagedPolicyName("AmazonSSMManagedInstanceCore")
  //   ]
  // });

  // const bastionHost = new ec2.Instance(stack, "BastionHost", {
  //   vpc,
  //   instanceType: new ec2.InstanceType("t3.micro"),
  //   machineImage: new ec2.AmazonLinuxImage(),
  //   vpcSubnets: { subnetType: ec2.SubnetType.PUBLIC },
  //   securityGroup: bastionSecurityGroup,
  //   role: bastionRole,
  //   keyPair: ec2.KeyPair.fromKeyPairName(stack, "ec2-bastion-key-pair", keyPair.keyName)
  // });

  // Create security group for bastion host
  // const bastionSecurityGroup = new ec2.SecurityGroup(stack, "bastion-sg", {
  //   vpc,
  //   description: "Security group for bastion host",
  //   allowAllOutbound: true
  // });
  //
  // // Allow SSH access to bastion host
  // bastionSecurityGroup.addIngressRule(ec2.Peer.anyIpv4(), ec2.Port.tcp(22), "Allow SSH access from anywhere");
  //
  // // Create bastion host
  // const bastionHost = new ec2.Instance(stack, "bastion-host", {
  //   vpc,
  //   vpcSubnets: { subnetType: ec2.SubnetType.PUBLIC },
  //   instanceType: ec2.InstanceType.of(ec2.InstanceClass.T3, ec2.InstanceSize.MICRO),
  //   machineImage: new ec2.AmazonLinuxImage(),
  //   securityGroup: bastionSecurityGroup,
  //   role: bastionRole,
  //   keyPair: ec2.KeyPair.fromKeyPairName(stack, "ec2-bastion-key-pair", keyPair.keyName)
  // });
  //
  // // Allow bastion host to access private subnets
  // vpc.privateSubnets.forEach((subnet, index) => {
  //   bastionSecurityGroup.addEgressRule(
  //     ec2.Peer.ipv4(subnet.ipv4CidrBlock),
  //     ec2.Port.allTcp(),
  //     `Allow access to private subnet ${index}`
  //   );
  // });
  //
  // vpc.publicSubnets.forEach((subnet, index) => {
  //   bastionSecurityGroup.addEgressRule(
  //     ec2.Peer.ipv4(subnet.ipv4CidrBlock),
  //     ec2.Port.allTcp(),
  //     `Allow access to private subnet ${index}`
  //   );
  // });

  // // Define the bastion host
  // const bastionSecurityGroup = new ec2.SecurityGroup(stack, "real-world-bastion-sg", {
  //   securityGroupName: "real-world-bastion-sg",
  //   vpc,
  //   description: "Allow SSH access",
  //   allowAllOutbound: true
  // });
  //
  // // Allow SSH access to the bastion host
  // bastionSecurityGroup.addIngressRule(ec2.Peer.anyIpv4(), ec2.Port.tcp(22), "Allow SSH access from anywhere");

  // const bastion = new ec2.BastionHostLinux(stack, "real-world-bastion-host", {
  //   vpc,
  //   subnetSelection: { subnetType: ec2.SubnetType.PUBLIC },
  //   instanceType: new ec2.InstanceType("t2.micro"),
  //   instanceName: "real-world-bastion-host",
  //   securityGroup: bastionSecurityGroup
  // });

  // bastion.instance.instance.addPropertyOverride("KeyName", keyPair.keyName);

  // Allow SSH access to the bastion host
  // bastion.connections.allowFromAnyIpv4(ec2.Port.tcp(22));
  // bastion.allowSshAccessFrom(ec2.Peer.anyIpv4());

  return {
    vpc: vpc,
    privateSubnets: vpc.selectSubnets({ subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS, onePerAz: true }).subnets,
    publicSubnets: vpc.selectSubnets({ subnetType: ec2.SubnetType.PUBLIC, onePerAz: true }).subnets,
    securityGroupId: vpc.vpcDefaultSecurityGroup
    // bastionSecurityGroupId: bastionSecurityGroup.securityGroupId
  };
}
