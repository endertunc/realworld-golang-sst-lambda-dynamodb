import type { App } from "sst/constructs";

// resource names are prefixed with stage and stack name to identify them easily
// during development, each developer gets their own stack by setting stage to their initials
export const getPrefixedResourceName = ({ stage, name }: App, resourceOrStackName?: string): string => {
  const prefix = `${stage}-${name}`;
  return resourceOrStackName ? `${prefix}-${resourceOrStackName}` : prefix;
};
