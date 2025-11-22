# Gemini Context: File Search Extension

Welcome to the File Search Extension! This extension provides a powerful tool, `query_knowledge_base`, which allows Gemini to find information within a dedicated knowledge base of documents you associate with your project.

The primary use case for this tool is to provide searchable access to supplementary context such as **hardware datasheets, design documents, API documentation, or other data sources** that are not suitable for being kept directly in the prompt context.

This document explains how you can use this extension to build and query a knowledge base for your own projects.

---

## How to Create a Searchable Knowledge Base for Your Project

To enable Gemini to answer questions using your supplementary documents, you need to follow a two-part process: **Setup** (to index your documents) and **Guidance** (to instruct the model).

### Part 1: Setup (One-time per project)

First, you must create a dedicated "File Search Store" for your project and upload your documents (e.g., PDFs, text files) into it. You can do this using the `file-search` CLI tool provided by this extension, or by using the MCP tools directly.

#### Using the CLI

1.  **Create a Store:** Open a terminal in your project's directory and run the following command, replacing `my-project-kb` with a unique name for your project's knowledge base.
    ```sh
    file-search store create my-project-kb
    ```

2.  **Upload Documents:** Use the `file upload` command to add your documents to the store.
    ```sh
    # Upload a specific PDF datasheet
    file-search file upload --store my-project-kb ./datasheets/component-a.pdf

    # Upload all documents from a directory
    file-search file upload --store my-project-kb ./docs/
    ```

### Part 2: Guidance (Instructing the Model)

Now that your knowledge base is indexed, you must tell the model how and when to use it for this project. You do this by creating a `GEMINI.md` file in the **root directory of your project**.

Below is a template you can copy and paste into your project's `GEMINI.md` file.

---

### Template for Your Project's `GEMINI.md`

```markdown
# Gemini Context: Project Knowledge Base for [Your Project Name]

This project is associated with a searchable knowledge base of supplementary documents (e.g., datasheets, design docs).

## File Search Tool

- **Tool Name:** `query_knowledge_base`
- **Project Knowledge Base Store:** `my-project-kb`  (IMPORTANT: Replace with your actual store name from Part 1)

## Instructions for the Model

1.  **Prioritize Knowledge Base Search:** When I ask a question about component specifications, architectural decisions, or information that would be found in external documentation, you **MUST** use the `query_knowledge_base` tool to find the answer before responding.

2.  **Use the Correct Store:** Always use the `store_name` parameter set to `my-project-kb` when calling the tool.

3.  **Formulate Effective Queries:** Your `query` parameter for the tool should be a concise question or a keyword search based on my prompt.

## Example Usage

-   **My Prompt:** "What is the maximum operating temperature for the T-800 series component?"
-   **Your Action:** (Call the tool) `query_knowledge_base(query='T-800 maximum operating temperature', store_name='my-project-kb')`

-   **My Prompt:** "Summarize the power-up sequence described in the 'System Architecture' document."
-   **Your Action:** (Call the tool) `query_knowledge_base(query='power-up sequence System Architecture', store_name='my-project-kb')`
```
