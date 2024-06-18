"""
translate from  https://github.com/zhufengme/GPTCommit/blob/main/gptcommit.sh

"""
import os
import subprocess
import requests

# 设置你的 OpenAI API 密钥
OPENAI_API_KEY = os.getenv('DEEPSEEK_API_KEY', '')
LLM_URL = "https://api.deepseek.com/v1/chat/completions"

# 设置你的 Proxy，默认使用HTTPS_PROXY环境变量
CURL_PROXY = os.getenv('HTTPS_PROXY', '')

def get_git_diff(diff_type):
    try:
        result = subprocess.check_output(['git', 'diff'] + diff_type, text=True)
        return result
    except subprocess.CalledProcessError as e:
        print(f"Error getting git diff: {e}")
        return ""

def generate_commit_message(diff):
    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {OPENAI_API_KEY}",
    }
    payload = {
        "model": "deepseek-coder",
        "messages": [
            {
                "role": "user",
                "content": f"分析以下代码更改并生成一个简洁的提交注释，只给出文本文字：\n\n{diff}\n\nCommit message:",
            }
        ],
        "max_tokens": 100,
        "temperature": 0.7,
    }
    proxies = {
        "https": CURL_PROXY,
    } if CURL_PROXY else {}

    try:
        response = requests.post(LLM_URL,
                                 headers=headers,
                                 proxies=proxies,
                                 json=payload,
                                 timeout=5)
        response.raise_for_status()
        return response.json()['choices'][0]['message']['content']
    except requests.RequestException as e:
        print(f"Error calling OpenAI API: {e}")
        return ""

def main():
    # 检查工作目录状态
    print("检查工作目录状态...")
    try:
        subprocess.run(['git', 'status'], check=True)
    except subprocess.CalledProcessError as e:
        print(f"Error checking git status: {e}")
        return

    # 获取工作目录和暂存区之间的差异
    working_diff = get_git_diff([])
    # 获取暂存区和HEAD之间的差异
    staged_diff = get_git_diff(['--cached'])

    # 合并差异
    diff = working_diff + staged_diff

    # 如果没有差异，退出
    if not diff.strip():
        print("没有发现差异。")
        return

    # 获取生成的提交注释
    commit_message = generate_commit_message(diff)
    print(commit_message)
    
    # git add
    print("git add...")
    try:
        subprocess.run(['git', 'add', '.',  '-A'], check=True)
    except subprocess.CalledProcessError as e:
        print(f"Error adding changes: {e}")
        return

    # 提交代码
    print("提交代码...")
    try:
        subprocess.run(['git', 'commit', '-m', commit_message], check=True)
    except subprocess.CalledProcessError as e:
        print(f"Error committing changes: {e}")
        return

if __name__ == "__main__":
    main()
