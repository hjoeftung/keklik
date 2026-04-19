import { useParams } from 'react-router-dom'

export default function InviteScreen() {
  const { token } = useParams()
  return <div>Accept Invite — token: {token}</div>
}
