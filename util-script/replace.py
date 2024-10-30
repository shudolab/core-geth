import os

def replace_in_file(file_path):
    # ファイルを読み込み
    with open(file_path, 'r', encoding='utf-8') as file:
        content = file.read()
    
    # 文字列を置換
    new_content = content.replace(
        'github.com/ethereum/go-ethereum', 
        'github.com/shudolab/core-geth'
    )
    
    # 変更があった場合のみファイルを書き込み
    if new_content != content:
        with open(file_path, 'w', encoding='utf-8') as file:
            file.write(new_content)
        print(f"Updated: {file_path}")

def process_directory(directory):
    # ディレクトリ内のすべてのファイルとフォルダを走査
    for root, dirs, files in os.walk(directory):
        # .goファイルのみを処理
        for file in files:
            if file.endswith('.go'):
                file_path = os.path.join(root, file)
                replace_in_file(file_path)

# カレントディレクトリから実行
process_directory('.')