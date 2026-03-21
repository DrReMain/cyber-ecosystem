import type { Tree } from '@nx/devkit';
import {
  generateFiles,
  joinPathFragments,
  normalizePath,
  logger,
} from '@nx/devkit';
import { join, dirname } from 'node:path';
import { existsSync } from 'node:fs';

/**
 * Finds the workspace root by searching for nx.json in parent directories.
 * This is more robust than hardcoding relative paths.
 */
function findWorkspaceRoot(startDir: string): string {
  let currentDir = startDir;

  while (currentDir !== dirname(currentDir)) {
    if (existsSync(join(currentDir, 'nx.json'))) {
      return currentDir;
    }
    currentDir = dirname(currentDir);
  }

  // Check the root directory as well
  if (existsSync(join(currentDir, 'nx.json'))) {
    return currentDir;
  }

  throw new Error('Could not find workspace root (directory containing nx.json)');
}

/**
 * Gets the list of existing Nx project names.
 */
function getExistingProjects(workspaceRoot: string): string[] {
  const { execSync } = require('node:child_process');
  const nxPath = join(workspaceRoot, 'nx');

  try {
    const output = execSync(`"${nxPath}" show projects`, {
      cwd: workspaceRoot,
      encoding: 'utf-8',
      stdio: ['pipe', 'pipe', 'pipe'],
    });

    return output.trim().split('\n').filter(Boolean);
  } catch (error) {
    logger.warn(`\n⚠️ Could not get project list from Nx`);
    return [];
  }
}

export interface KratosBaseGeneratorSchema {
  /** The name of the Kratos service project */
  name: string;
  /** The directory where the project will be generated (default: 'services') */
  directory?: string;
  /** Comma-separated list of tags for the project */
  tags?: string;
}

/**
 * Validates the project name format.
 * Must start with a lowercase letter and contain only lowercase letters, numbers, and hyphens.
 */
function validateProjectName(name: string): void {
  const namePattern = /^[a-z][a-z0-9-]*$/;
  if (!namePattern.test(name)) {
    throw new Error(
      `Invalid project name "${name}". ` +
      `Name must start with a lowercase letter and contain only lowercase letters, numbers, and hyphens.`
    );
  }
}

/**
 * Normalizes the directory path and ensures it doesn't have trailing slashes.
 */
function normalizeDirectory(directory: string | undefined): string {
  const defaultDirectory = 'services';
  const dir = (directory || defaultDirectory).trim();

  if (!dir) {
    return defaultDirectory;
  }

  // Remove leading and trailing slashes, then normalize
  return normalizePath(dir).replace(/^\/+|\/+$/g, '');
}

/**
 * Calculates the relative path from the project root to the workspace root.
 * This is used for the $schema property in project.json.
 */
function calculateRelativePath(projectRoot: string): string {
  const depth = projectRoot.split('/').length;
  // Remove trailing slash to avoid double slashes in the path
  return '../'.repeat(depth).replace(/\/$/, '');
}

/**
 * Updates buf.yaml to add a new module entry (only path field).
 */
function updateBufYaml(tree: Tree, projectRoot: string, name: string): void {
  const bufYamlPath = 'buf.yaml';
  const content = tree.read(bufYamlPath);

  if (!content) {
    logger.warn(`\n⚠️ Could not find ${bufYamlPath}, skipping module registration`);
    return;
  }

  const text = content.toString();
  const lines = text.split('\n');

  // Find the modules section and the last module entry
  let modulesStartIndex = -1;
  let lastModuleIndex = -1;

  for (let i = 0; i < lines.length; i++) {
    if (lines[i].trim().startsWith('modules:')) {
      modulesStartIndex = i;
    }
    if (modulesStartIndex !== -1 && lines[i].trim().startsWith('- path:')) {
      lastModuleIndex = i;
    }
    // Stop when we reach another section (like deps:)
    if (modulesStartIndex !== -1 && lastModuleIndex !== -1 &&
        lines[i].trim() && !lines[i].trim().startsWith('-') &&
        !lines[i].trim().startsWith('#') && !lines[i].trim().startsWith('path:')) {
      break;
    }
  }

  if (modulesStartIndex === -1) {
    logger.warn(`\n⚠️ Could not find modules section in ${bufYamlPath}, skipping module registration`);
    return;
  }

  // Check if module already exists
  const modulePath = `${projectRoot}/api`;
  if (text.includes(`path: ${modulePath}`)) {
    logger.info(`\nℹ️ Module ${modulePath} already exists in ${bufYamlPath}`);
    return;
  }

  // Create new module entry (only path, no name)
  const newModule = `  - path: ${modulePath}`;

  // Insert after the last module entry
  if (lastModuleIndex !== -1) {
    lines.splice(lastModuleIndex + 1, 0, newModule);
  } else {
    // No existing modules, add after modules: line
    lines.splice(modulesStartIndex + 1, 0, newModule);
  }

  tree.write(bufYamlPath, lines.join('\n'));
  logger.info(`\n📝 Added module path: ${modulePath} to ${bufYamlPath}`);
}

export default async function kratosBaseGenerator(
  tree: Tree,
  options: KratosBaseGeneratorSchema
): Promise<() => Promise<void>> {
  const { name, tags } = options;

  // Validate project name
  validateProjectName(name);

  // Normalize directory path
  const directory = normalizeDirectory(options.directory);

  // Calculate the full project path
  const projectRoot = joinPathFragments(directory, name);

  // Find workspace root and check if project name already exists
  const workspaceRoot = findWorkspaceRoot(__dirname);
  const existingProjects = getExistingProjects(workspaceRoot);

  // Project name is just the name (e.g., "test-service")
  const projectName = name;

  if (existingProjects.includes(projectName)) {
    throw new Error(
      `Project "${projectName}" already exists in the workspace. ` +
      `Please choose a different name.`
    );
  }

  // Check if project directory already exists
  if (existsSync(join(workspaceRoot, projectRoot))) {
    throw new Error(
      `Directory "${projectRoot}" already exists. ` +
      `Please choose a different name or directory.`
    );
  }
  const relativePath = calculateRelativePath(projectRoot);

  // Calculate paths for generated code
  // genGoPath: relative path from project to gen/go/{name}
  // genOasPath: relative path from project to gen/oas/{name}
  const genGoPath = joinPathFragments(relativePath, 'gen', 'go', name);
  const genOasPath = joinPathFragments(relativePath, 'gen', 'oas', name);

  // Generate database name (replace hyphens with underscores)
  const dbName = name.replace(/-/g, '_');

  // Generate proto package name (replace hyphens with underscores)
  const protoPackage = name.replace(/-/g, '_');

  // Generate Go package name (camelCase: test-service -> testService)
  const goPackage = name
    .split('-')
    .map((part, index) =>
      index === 0 ? part : part.charAt(0).toUpperCase() + part.slice(1)
    )
    .join('');

  // Parse tags into an array
  const parsedTags = tags
    ? tags.split(',').map((tag) => tag.trim()).filter(Boolean)
    : [];

  // Generate files from templates
  const templatePath = join(__dirname, 'files');
  generateFiles(tree, templatePath, projectRoot, {
    name,
    directory,
    projectName,
    projectRoot,
    relativePath,
    genGoPath,
    genOasPath,
    dbName,
    protoPackage,
    goPackage,
    tags: parsedTags.join(','),
    tmpl: '',
  });

  // Update buf.yaml to add the new module
  updateBufYaml(tree, projectRoot, name);

  logger.info(`\n✅ Successfully generated Kratos project "${name}" at "${projectRoot}"`);

  return async () => {
    const { execSync } = await import('node:child_process');
    // Find workspace root by looking for nx.json
    const workspaceRoot = findWorkspaceRoot(__dirname);
    const nxPath = join(workspaceRoot, 'nx');

    // Run proto
    logger.info(`\n🔄 Running [proto] for project "${projectName}"...`);
    try {
      execSync(`"${nxPath}" run ${projectName}:proto`, {
        cwd: workspaceRoot,
        stdio: 'inherit',
      });
      logger.info(`\n✅ Successfully ran [proto] for project "${projectName}"`);
    } catch (error) {
      logger.error(`\n❌ Failed to run [proto] for project "${projectName}"`);
      throw error;
    }

    // Run generate
    logger.info(`\n🔄 Running [generate] for project "${projectName}"...`);
    try {
      execSync(`"${nxPath}" run ${projectName}:generate`, {
        cwd: workspaceRoot,
        stdio: 'inherit',
      });
      logger.info(`\n✅ Successfully ran [generate] for project "${projectName}"`);
    } catch (error) {
      logger.error(`\n❌ Failed to run [generate] for project "${projectName}"`);
      throw error;
    }
  };
}
