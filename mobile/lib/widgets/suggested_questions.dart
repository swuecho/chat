import 'package:flutter/material.dart';

class SuggestedQuestions extends StatelessWidget {
  const SuggestedQuestions({
    super.key,
    required this.questions,
    required this.loading,
    required this.onSelect,
    required this.onGenerateMore,
    required this.generating,
    required this.batches,
    required this.currentBatch,
    required this.onPreviousBatch,
    required this.onNextBatch,
  });

  final List<String> questions;
  final bool loading;
  final ValueChanged<String> onSelect;
  final VoidCallback onGenerateMore;
  final bool generating;
  final List<List<String>> batches;
  final int currentBatch;
  final VoidCallback onPreviousBatch;
  final VoidCallback onNextBatch;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final hasPreviousBatch = batches.length > 1 && currentBatch > 0;
    final hasNextBatch = batches.length > 1 && currentBatch < batches.length - 1;
    return Container(
      margin: const EdgeInsets.only(top: 8),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: const Color(0xFFF8FAFC),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: const Color(0xFFE2E8F0)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.lightbulb_outline, size: 16, color: Color(0xFF2563EB)),
              const SizedBox(width: 6),
              Text(
                'Suggested questions',
                style: theme.textTheme.labelMedium,
              ),
              if (loading) ...[
                const SizedBox(width: 8),
                const SizedBox(
                  height: 12,
                  width: 12,
                  child: CircularProgressIndicator(strokeWidth: 2),
                ),
              ],
              if (generating) ...[
                const SizedBox(width: 8),
                const SizedBox(
                  height: 12,
                  width: 12,
                  child: CircularProgressIndicator(strokeWidth: 2),
                ),
              ],
              const Spacer(),
              if (batches.length > 1)
                Row(
                  children: [
                    Text(
                      '${currentBatch + 1}/${batches.length}',
                      style: theme.textTheme.labelSmall,
                    ),
                    IconButton(
                      icon: const Icon(Icons.chevron_left, size: 18),
                      onPressed: hasPreviousBatch ? onPreviousBatch : null,
                    ),
                    IconButton(
                      icon: const Icon(Icons.chevron_right, size: 18),
                      onPressed: hasNextBatch ? onNextBatch : null,
                    ),
                  ],
                ),
            ],
          ),
          if (!loading && questions.isNotEmpty) ...[
            const SizedBox(height: 8),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                for (final question in questions)
                  InkWell(
                    borderRadius: BorderRadius.circular(12),
                    onTap: () => onSelect(question),
                    child: Container(
                      constraints: const BoxConstraints(maxWidth: 320),
                      padding: const EdgeInsets.symmetric(
                        horizontal: 12,
                        vertical: 8,
                      ),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(color: const Color(0xFFE2E8F0)),
                      ),
                      child: Text(
                        question,
                        softWrap: true,
                        style: theme.textTheme.bodySmall,
                      ),
                    ),
                  ),
              ],
            ),
            const SizedBox(height: 8),
            Align(
              alignment: Alignment.centerLeft,
              child: TextButton.icon(
                onPressed: generating ? null : onGenerateMore,
                icon: const Icon(Icons.refresh, size: 16),
                label: Text(generating ? 'Generating...' : 'Generate more'),
              ),
            ),
          ],
        ],
      ),
    );
  }
}
