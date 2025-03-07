import subprocess
from datetime import datetime

def get_local_branches():
    # Get the list of local branches
    result = subprocess.run(['git', 'branch'], stdout=subprocess.PIPE, text=True)
    branches = result.stdout.splitlines()
    # Remove the '*' from the current branch
    branches = [branch.strip('*').strip() for branch in branches]
    return branches

def get_branch_last_commit_date(branch):
    # Get the last commit date of the branch
    result = subprocess.run(['git', 'log', '-1', '--format=%cd', '--date=iso', branch], stdout=subprocess.PIPE, text=True)
    date_str = result.stdout.strip()
    return datetime.strptime(date_str, '%Y-%m-%d %H:%M:%S %z')

def delete_branch(branch):
    # Delete the branch
    subprocess.run(['git', 'branch', '-D', branch])

def confirm_deletion(branch):
    # Ask the user to confirm deletion
    response = input(f"Do you want to delete the branch '{branch}'? (y/n): ").strip().lower()
    return response == 'y'

def main():
    # Get all local branches
    branches = get_local_branches()

    # Get the last commit date for each branch
    branch_dates = [(branch, get_branch_last_commit_date(branch)) for branch in branches]

    # Sort branches by last commit date (oldest first)
    branch_dates.sort(key=lambda x: x[1])

    # Get the oldest 5 branches
    oldest_branches = [branch for branch, _ in branch_dates[:5]]

    # Delete the oldest 5 branches with confirmation
    for branch in oldest_branches:
        if confirm_deletion(branch):
            print(f"Deleting branch: {branch}")
            delete_branch(branch)
        else:
            print(f"Skipping branch: {branch}")

if __name__ == "__main__":
    main()
