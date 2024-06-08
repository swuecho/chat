import subprocess
import time

# Get the list of merged branches
output = subprocess.check_output(['git', 'branch']).decode('utf-8')
print(output)
# Exclude the current branch and any branches named "main", "master", or "develop"
exclude_branches = ['main', 'master', 'develop']
branches = [line.strip('* ') for line in output.split('\n') if line.strip('* ') not in exclude_branches and line.strip()]
print(branches)
# Get the current time one month ago
one_month_ago = time.time() - 30 * 24 * 60 * 60

# Delete branches that have not been updated in the last month
for branch in branches:
    # Get the Unix timestamp of the last commit on the branch
    output = subprocess.check_output(['git', 'log', '-1', '--format=%at', branch]).decode('utf-8')
    print(output)
    last_commit_time = int(output.strip())

    # Delete the branch if its last commit is older than one month
    if last_commit_time < one_month_ago:
        subprocess.call(['git', 'branch', '-D', branch])