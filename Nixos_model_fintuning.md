# 🦙 Fine-Tuning Llama 3/4 for NixOS: Detailed Project Plan

This document provides a comprehensive, step-by-step guide to fine-tuning a Llama 3 or Llama 4 Large Language Model (LLM) for NixOS-specific tasks. The goal is to create a model that excels at answering NixOS questions, generating config snippets, and assisting with NixOS system management.

---

## Prerequisites

- **Hardware:**

  - Access to a machine with a modern GPU (NVIDIA 24GB+ VRAM recommended) or sufficient CPU/RAM for smaller models.
  - At least 100GB free disk space for datasets, checkpoints, and model weights.

- **Software:**

  - Linux (recommended), macOS, or WSL2.
  - Python 3.9+ and pip.
  - [Ollama](https://github.com/ollama/ollama) (for local fine-tuning) **or** [Hugging Face Transformers](https://huggingface.co/docs/transformers/installation).
  - [git](https://git-scm.com/), [jq](https://stedolan.github.io/jq/), [curl](https://curl.se/).
  - (Optional) Docker for containerized workflows.

- **Accounts/Access:**

  - Access to Llama 3/4 weights (Meta registration may be required).
  - (Optional) Hugging Face account for model hosting.

---

## Step-by-Step Process

### 1. Define Objectives

- Clarify the use cases: config generation, troubleshooting, option explanations, packaging, etc.
- Decide on the base model (Llama 3 or 4) and deployment method (Ollama, HF, etc.).

---

### 2. Data Collection

- **Documentation:**

  - Download NixOS and Home Manager manuals (HTML, Markdown, JSON).
  - Use scripts to scrape the NixOS Wiki and option search.

- **Real Configs:**

  - Collect `configuration.nix`, `home.nix`, and flake files from public repos (GitHub search: `filename:configuration.nix`).
  - Anonymize any sensitive data.

- **Community Q&A:**

  - Export relevant threads from Discourse, Stack Overflow, and GitHub Issues.

- **Logs & Errors:**

  - Gather real-world NixOS error logs and troubleshooting sessions.

- **AI-Generated Data:**

  - Use existing nixai outputs as additional training data.

---

### 3. Data Cleaning & Formatting

- **Deduplication:**

  Remove duplicate or near-duplicate entries.

- **Formatting:**

  Convert all data to a unified format (Alpaca-style or OpenAI JSONL):

  ```json
  {"instruction": "How do I enable SSH in NixOS?", "input": "", "output": "Add services.openssh.enable = true; to your configuration.nix and rebuild."}
  ```

  Ensure each entry is a clear prompt/response pair.

- **Validation:**

  - Manually review a sample for correctness and NixOS idioms.
  - Ensure a mix of basic, common, and advanced examples.
  - Include troubleshooting and error resolution cases.

---

### 4. Environment Setup

- **Install dependencies:**

  - For Ollama: `curl -fsSL https://ollama.com/install.sh | sh`
  - For Hugging Face: `pip install transformers datasets trl` (and CUDA toolkit if using GPU)

- **Download base model weights:**

  - Register and download Llama 3/4 weights from Meta or Hugging Face.
  - Place weights in the appropriate directory for your tool.

- **Prepare training scripts/configs:**

  - For Ollama: create a `Modelfile`.
  - For HF: prepare a training config or use `run_sft.py`.

---

### 5. Fine-Tuning

- **Ollama:**

  1. Place your dataset (e.g., `data.jsonl`) and `Modelfile` in the working directory.
  2. Example `Modelfile`:

     ```Dockerfile
     FROM llama3
     TRAIN data.jsonl
     ```

  3. Run:

     ```sh
     ollama create nixos-llama3 -f Modelfile
     ```

  4. Monitor logs for progress and errors.

- **Hugging Face:**

  1. Prepare your dataset as JSONL or CSV.
  2. Use a script like:

     ```sh
     python run_sft.py \
       --model_name_or_path /path/to/llama3 \
       --train_file data.jsonl \
       --output_dir ./nixos-llama3-finetuned \
       --per_device_train_batch_size 2 \
       --num_train_epochs 3
     ```

  3. Adjust hyperparameters as needed for your hardware.

---

### 6. Evaluation

- **Create a test set** of NixOS-specific prompts and expected outputs.
- **Run the model** on the test set and compare outputs for:

  - Correctness of config snippets.
  - Quality of troubleshooting and explanations.
  - Coverage of advanced and edge cases.

- **Iterate:**

  - Add more data for weak areas.
  - Refine prompts and outputs.
  - Repeat fine-tuning as needed.

---

### 7. Integration & Deployment

- **Ollama:**

  - Update `configs/default.yaml` in nixai to use your custom model:

    ```yaml
    ai_provider: ollama
    ai_model: nixos-llama3
    ```

- **Hugging Face:**

  - Serve the model locally or via API and update nixai config accordingly.

- **Test:**

  - Run `nixai` with the new model and verify improved NixOS support.

---

### 8. Documentation & Sharing

- Document the process, dataset sources, and any scripts used.
- (If license allows) Share the fine-tuned model and instructions for others to use it.

---

## Project Checklist

- [ ] Hardware and software prerequisites met
- [ ] Objectives and use cases defined
- [ ] Data collected from all relevant sources
- [ ] Data cleaned, formatted, and validated
- [ ] Environment and dependencies set up
- [ ] Base model weights downloaded
- [ ] Fine-tuning run and monitored
- [ ] Model evaluated and iterated
- [ ] Integrated with nixai and tested
- [ ] Documentation written and (optionally) model shared

---

**References:**

- [Ollama Fine-Tuning Guide](https://github.com/ollama/ollama/blob/main/docs/fine-tune.md)
- [Hugging Face Transformers Fine-Tuning](https://huggingface.co/docs/transformers/training)
- [NixOS Manual](https://nixos.org/manual/nixos/stable/)
- [Home Manager Manual](https://nix-community.github.io/home-manager/)

---

> Use this checklist and step-by-step guide for any future NixOS LLM fine-tuning projects.
