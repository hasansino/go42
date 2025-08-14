#!/usr/bin/env node

const fs = require('fs-extra');
const path = require('path');
const glob = require('glob');

// Define source directories and their target locations in Docusaurus
const docSources = [
  {
    source: '../core',
    target: 'docs/core',
    label: 'Core Documentation',
    description: 'Core project documentation',
  },
  {
    source: '../adr',
    target: 'docs/architecture',
    label: 'Architecture Decisions',
    description: 'Architecture Decision Records (ADRs)',
  },
  {
    source: '../brd',
    target: 'docs/business',
    label: 'Business Requirements',
    description: 'Business Requirement Documents (BRDs)',
  },
];

// Ensure target directories exist and are clean
console.log('🧹 Cleaning existing gathered docs...');
docSources.forEach(({ target }) => {
  const targetPath = path.join(__dirname, target);
  if (fs.existsSync(targetPath)) {
    fs.removeSync(targetPath);
  }
  fs.ensureDirSync(targetPath);
});

// Copy documentation from each source
docSources.forEach(({ source, target, label, description }) => {
  const sourcePath = path.join(__dirname, source);
  const targetPath = path.join(__dirname, target);
  
  console.log(`📚 Gathering ${label} from ${source}...`);
  
  // Check if source directory exists
  if (!fs.existsSync(sourcePath)) {
    console.log(`  ⚠️  Source directory ${source} does not exist, creating placeholder...`);
    return;
  }
  
  // Find all markdown files in source directory
  const files = glob.sync('**/*.md', { 
    cwd: sourcePath,
    nodir: true 
  });
  
  if (files.length === 0) {
    console.log(`  ⚠️  No markdown files found in ${source}, creating placeholder...`);
    return;
  }
  
  // Copy each file and add frontmatter if needed
  files.forEach(file => {
    const sourceFile = path.join(sourcePath, file);
    const targetFile = path.join(targetPath, file);
    
    // Ensure target subdirectory exists
    fs.ensureDirSync(path.dirname(targetFile));
    
    // Read file content
    let content = fs.readFileSync(sourceFile, 'utf8');
    
    // Add frontmatter if it doesn't exist
    if (!content.startsWith('---')) {
      const fileName = path.basename(file, '.md');
      const title = fileName
        .replace(/[-_]/g, ' ')
        .replace(/\b\w/g, char => char.toUpperCase());
      
      const frontmatter = `---
id: ${fileName}
title: ${title}
---

`;
      content = frontmatter + content;
    }
    
    // Write to target
    fs.writeFileSync(targetFile, content);
    console.log(`  ✅ Copied ${file}`);
  });
  
  // Create category file for better sidebar organization
  const categoryFile = path.join(targetPath, '_category_.json');
  const categoryContent = {
    label: label,
    position: docSources.findIndex(s => s.target === target) + 2, // Start from position 2
    collapsed: true, // Categories collapsed by default
    link: {
      type: 'generated-index',
      description: description,
    },
  };
  
  fs.writeJsonSync(categoryFile, categoryContent, { spaces: 2 });
  console.log(`  ✅ Created category configuration`);
});

console.log('\n✨ Documentation gathering complete!');
