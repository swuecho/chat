import { displayLocaleDate, formatYearMonth } from '@/utils/date'





export function generateAPIHelper(uuid: string, apiToken: string, origin: string) {
        const data = {
                "message": "Your message here",
                "snapshot_uuid": uuid,
                "stream": false,
        }
        return `curl -X POST ${origin}/api/chatbot -H "Content-Type: application/json" -H "Authorization: Bearer ${apiToken}" -d '${JSON.stringify(data)}'`
}

export function getChatbotPosts(posts: Snapshot.Snapshot[]) {
        return posts
                .filter((post: Snapshot.Snapshot) => post.typ === 'chatbot')
                .map((post: Snapshot.Snapshot): Snapshot.PostLink => ({
                        uuid: post.uuid,
                        date: displayLocaleDate(post.createdAt),
                        title: post.title,
                }))
}

export function getSnapshotPosts(posts: Snapshot.Snapshot[]) {
        return posts
                .filter((post: Snapshot.Snapshot) => post.typ === 'snapshot')
                .map((post: Snapshot.Snapshot): Snapshot.PostLink => ({
                        uuid: post.uuid,
                        date: displayLocaleDate(post.createdAt),
                        title: post.title,
                }))
}

export function postsByYearMonthTransform(posts: Snapshot.PostLink[]) {
        const init: Record<string, Snapshot.PostLink[]> = {}
        return posts.reduce((acc, post) => {
                const yearMonth = formatYearMonth(new Date(post.date))
                if (!acc[yearMonth])
                        acc[yearMonth] = []

                acc[yearMonth].push(post)
                return acc
        }, init)
}

export function getSnapshotPostLinks(snapshots: Snapshot.Snapshot[]): Record<string, Snapshot.PostLink[]> {
        const snapshotPosts = getSnapshotPosts(snapshots)
        return postsByYearMonthTransform(snapshotPosts)
}

export function getBotPostLinks(bots: Snapshot.Snapshot[]): Record<string, Snapshot.PostLink[]> {
        const chatbotPosts = getChatbotPosts(bots)
        return postsByYearMonthTransform(chatbotPosts)
}