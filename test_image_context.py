"""
测试生图模型是否具有上下文能力。
双击运行，按提示输入 API 信息即可。
"""

import requests
import base64
import time
import json
from pathlib import Path


def generate_image(base_url: str, api_key: str, model: str, prompt: str, size: str = "1792x1024") -> dict:
    url = f"{base_url.rstrip('/')}/v1/images/generations"
    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {api_key}",
    }
    payload = {
        "model": model,
        "prompt": prompt,
        "n": 1,
        "size": size,
    }

    print(f"\n{'='*60}")
    print(f"Prompt: {prompt[:80]}...")
    print(f"{'='*60}")
    print("请求中，可能需要 30-60 秒...")

    resp = requests.post(url, headers=headers, json=payload, timeout=180)

    if resp.status_code != 200:
        print(f"[ERROR] status {resp.status_code}")
        print(resp.text[:500])
        return None

    data = resp.json()
    print("生成成功!")
    return data


def save_image(data: dict, filename: str) -> str:
    if not data or not data.get("data"):
        return None

    item = data["data"][0]
    output_dir = Path("test_output")
    output_dir.mkdir(exist_ok=True)

    filepath = output_dir / filename

    if item.get("b64_json"):
        img_bytes = base64.b64decode(item["b64_json"])
        filepath.write_bytes(img_bytes)
    elif item.get("url"):
        print("下载图片中...")
        img_resp = requests.get(item["url"], timeout=60)
        filepath.write_bytes(img_resp.content)
    else:
        print("响应中没有图片数据")
        return None

    revised = item.get("revised_prompt", "")
    if revised:
        print(f"模型修改后的 prompt: {revised[:100]}")

    print(f"已保存: {filepath}")
    return str(filepath)


def main():
    print("=" * 60)
    print("  生图模型上下文能力测试")
    print("=" * 60)
    print()
    print("本测试会生成 3 张图来判断模型是否有跨请求的记忆能力。")
    print("预计消耗 3 次生图额度，耗时 2-3 分钟。")
    print()

    # 交互式输入
    api_key = input("请输入 API Key: ").strip()
    if not api_key:
        print("API Key 不能为空!")
        input("按回车退出...")
        return

    base_url = input("请输入 API Base URL (直接回车默认 https://api.openai.com): ").strip()
    if not base_url:
        base_url = "https://api.openai.com"

    model = input("请输入模型名称 (直接回车默认 gpt-image-2): ").strip()
    if not model:
        model = "gpt-image-2"

    size = input("请输入图片尺寸 (直接回车默认 1792x1024): ").strip()
    if not size:
        size = "1792x1024"

    print()
    print(f"  API: {base_url}")
    print(f"  模型: {model}")
    print(f"  尺寸: {size}")
    print()
    confirm = input("确认开始测试? (y/n): ").strip().lower()
    if confirm not in ("y", "yes", ""):
        print("已取消。")
        input("按回车退出...")
        return

    # ===== 测试开始 =====

    prompt1 = (
        "A white cat wearing a red knitted scarf, sitting in a snowy forest. "
        "The cat has bright green eyes and a small black spot on its left ear. "
        "Illustration style, soft lighting."
    )

    prompt2 = (
        "The same cat from before, now sitting on a cozy brown leather sofa "
        "in a warm living room with a fireplace. Same illustration style."
    )

    prompt3 = (
        "A cat sitting on a cozy brown leather sofa in a warm living room "
        "with a fireplace. Illustration style, soft lighting."
    )

    print("\n" + "=" * 60)
    print("[1/3] 生成第一张图（建立角色：白猫+红围巾+绿眼+左耳黑点）")
    result1 = generate_image(base_url, api_key, model, prompt1, size)
    save_image(result1, "01_establish_character.png")

    time.sleep(2)

    print("\n" + "=" * 60)
    print("[2/3] 生成第二张图（引用'the same cat from before'，不重复外观）")
    result2 = generate_image(base_url, api_key, model, prompt2, size)
    save_image(result2, "02_reference_same_cat.png")

    time.sleep(2)

    print("\n" + "=" * 60)
    print("[3/3] 生成对照图（独立描述，不引用前文）")
    result3 = generate_image(base_url, api_key, model, prompt3, size)
    save_image(result3, "03_control_independent.png")

    # ===== 结果分析 =====

    print("\n" + "=" * 60)
    print("  测试完成!")
    print("=" * 60)
    print()
    print("请打开 test_output 文件夹对比三张图：")
    print()
    print("  01_establish_character.png")
    print("     → 白猫 + 红围巾 + 绿眼睛 + 左耳黑点 + 雪地")
    print()
    print("  02_reference_same_cat.png")
    print("     → prompt 只说了 'the same cat from before'")
    print("     → 如果有上下文：应保留红围巾/绿眼/黑点")
    print("     → 如果无上下文：猫的外观会是随机的")
    print()
    print("  03_control_independent.png")
    print("     → 对照组，完全独立的 prompt，猫外观随机")
    print()
    print("-" * 60)
    print("判断标准：")
    print("  图02 保留了红围巾+绿眼+黑点 → 模型有上下文记忆")
    print("  图02 和图03 的猫都是随机外观 → 模型无上下文")
    print("  图02 是白猫但没围巾/黑点   → 模型从'same cat'猜测，无真正记忆")
    print("-" * 60)

    # 保存响应记录
    results = {
        "test_time": time.strftime("%Y-%m-%d %H:%M:%S"),
        "model": model,
        "base_url": base_url,
        "size": size,
        "prompts": [prompt1, prompt2, prompt3],
        "responses": [
            result1 if result1 else "FAILED",
            result2 if result2 else "FAILED",
            result3 if result3 else "FAILED",
        ],
    }

    for r in results["responses"]:
        if isinstance(r, dict) and r.get("data"):
            for item in r["data"]:
                if item.get("b64_json"):
                    item["b64_json"] = f"[BASE64_DATA, {len(item['b64_json'])} chars]"

    output_dir = Path("test_output")
    output_dir.mkdir(exist_ok=True)
    output_path = output_dir / "results.json"
    output_path.write_text(json.dumps(results, indent=2, ensure_ascii=False))
    print(f"\nAPI 响应详情已保存到: {output_path}")

    print()
    input("按回车退出...")


if __name__ == "__main__":
    try:
        main()
    except Exception as e:
        print(f"\n发生错误: {e}")
        input("按回车退出...")
