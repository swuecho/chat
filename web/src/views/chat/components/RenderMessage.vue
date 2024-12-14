<script lang="ts">
import type { MessageRenderMessage } from 'naive-ui'
import { NAlert } from 'naive-ui'
import { h } from 'vue'

const renderMessage: MessageRenderMessage = (props) => {
        const { type } = props
        return h(
                NAlert,
                {
                        closable: props.closable,
                        onClose: props.onClose,
                        type: type === 'loading' ? 'default' : type,
                        style: {
                                boxShadow: 'var(--n-box-shadow)',
                                maxWidth: 'calc(100vw - 32px)',
                                width: '480px'
                        }
                },
                {
                        default: () => props.content
                }
        )
}

export default renderMessage;
</script>