import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_markdown/flutter_markdown.dart';

class ThinkingSection extends StatefulWidget {
  const ThinkingSection({
    super.key,
    required this.content,
    required this.styleSheet,
    this.defaultExpanded = true,
    this.maxLines = 20,
  });

  final String content;
  final MarkdownStyleSheet styleSheet;
  final bool defaultExpanded;
  final int maxLines;

  @override
  State<ThinkingSection> createState() => _ThinkingSectionState();
}

class _ThinkingSectionState extends State<ThinkingSection> {
  late bool _isExpanded;
  bool _isCopied = false;

  @override
  void initState() {
    super.initState();
    _isExpanded = widget.defaultExpanded;
  }

  bool get _isCollapsible {
    return widget.content.split('\n').length > widget.maxLines;
  }

  String get _collapsedContent {
    return widget.content.split('\n').take(widget.maxLines).join('\n').trimRight();
  }

  Future<void> _copyContent() async {
    await Clipboard.setData(ClipboardData(text: widget.content));
    if (!mounted) return;
    setState(() => _isCopied = true);
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('Thinking copied to clipboard'),
        duration: Duration(seconds: 2),
      ),
    );
    Future.delayed(const Duration(seconds: 2), () {
      if (mounted) {
        setState(() => _isCopied = false);
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final bodyContent = _isExpanded || !_isCollapsible
        ? widget.content
        : _collapsedContent;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      margin: const EdgeInsets.only(bottom: 8),
      decoration: BoxDecoration(
        color: const Color(0xFFF8FAFC),
        borderRadius: BorderRadius.circular(12),
        border: const Border(
          left: BorderSide(color: Color(0xFF84CC16), width: 3),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Text(
                '💭 Thinking',
                style: theme.textTheme.labelLarge?.copyWith(
                  color: const Color(0xFF475569),
                  fontWeight: FontWeight.w600,
                ),
              ),
              const Spacer(),
              IconButton(
                iconSize: 18,
                visualDensity: VisualDensity.compact,
                padding: EdgeInsets.zero,
                icon: Icon(
                  _isCopied ? Icons.check : Icons.copy,
                  color: _isCopied
                      ? const Color(0xFF16A34A)
                      : const Color(0xFF64748B),
                ),
                tooltip: _isCopied ? 'Copied' : 'Copy thinking',
                onPressed: widget.content.trim().isEmpty ? null : _copyContent,
              ),
              if (_isCollapsible)
                IconButton(
                  iconSize: 18,
                  visualDensity: VisualDensity.compact,
                  padding: EdgeInsets.zero,
                  icon: Icon(
                    _isExpanded ? Icons.expand_less : Icons.expand_more,
                    color: const Color(0xFF64748B),
                  ),
                  tooltip: _isExpanded ? 'Collapse thinking' : 'Expand thinking',
                  onPressed: () => setState(() => _isExpanded = !_isExpanded),
                ),
            ],
          ),
          if (bodyContent.trim().isNotEmpty)
            MarkdownBody(
              data: bodyContent,
              selectable: true,
              softLineBreak: true,
              styleSheet: widget.styleSheet,
            ),
          if (_isCollapsible && !_isExpanded)
            Padding(
              padding: const EdgeInsets.only(top: 6),
              child: TextButton(
                onPressed: () => setState(() => _isExpanded = true),
                child: const Text('Show more thinking'),
              ),
            ),
        ],
      ),
    );
  }
}
