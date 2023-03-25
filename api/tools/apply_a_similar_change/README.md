idea: sometime, a change is very similar to a previous change, such as another field to a table.
that will also need to add a field to the struct represent that table. etc.

1. get the changeset of previous change when add a field A with type Ta
2. let gpt generate a new changeset for add field B with type Tb based on the changeset of adding A.
3. apply the changeset to the git repo.

=== GPT polished version

Have you ever had to make a change to a project that was very similar to a previous change? For example, adding another field to a table in a database or adding another parameter to a function? If so, you know that it can be time-consuming to make these changes, especially if you have to make them manually each time.

Luckily, there is a solution to this problem: using GPT (Generative Pre-trained Transformer) to generate a new changeset based on a previous one.

Here's how it works:

1. Get the changeset of the previous change: When you made the previous change, you should have created a Git commit with a changeset that describes what you did. You can use Git to retrieve this changeset.

2. Use GPT to generate a new changeset based on the previous one: GPT is a deep learning model that can generate text based on an input prompt. In this case, you can use the previous changeset as the input prompt for GPT and ask it to generate a new changeset that adds the new field or parameter. GPT will generate a new changeset that is similar to the previous one but with the new field or parameter added.

3. Apply the new changeset to the Git repo: Once you have the new changeset, you can apply it to the Git repo using Git. This will automatically make the necessary changes to your project without you having to do it manually.

Using GPT to generate changesets can save you a lot of time and effort when making similar changes to your project. Instead of having to manually make the same changes over and over again, you can use GPT to generate a new changeset in just a few minutes. So the next time you need to make a similar change to your project, consider using GPT to help you out
