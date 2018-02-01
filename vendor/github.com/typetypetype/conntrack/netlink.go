package conntrack

import (
	"errors"
	"syscall"
)

// NFNL_MSG_TYPE
func nflnMsgType(x uint16) uint8 {
	return uint8(x & 0x00ff)
}

// NFNL_SUBSYS_ID
func nflnSubsysID(x uint16) uint8 {
	return uint8((x & 0xff00) >> 8)
}

// from src/libnfnetlink.c
func nfnlIsError(hdr syscall.NlMsghdr) error {
	if hdr.Type == syscall.NLMSG_ERROR {
		return errors.New("NLMSG_ERROR")
	}
	if hdr.Type == syscall.NLMSG_DONE && hdr.Flags&syscall.NLM_F_MULTI > 0 {
		return errors.New("Done!")
	}
	return nil
	/*
	   // This message is an ACK or a DONE
	   if (nlh->nlmsg_type == NLMSG_ERROR ||
	       (nlh->nlmsg_type == NLMSG_DONE &&
	       nlh->nlmsg_flags & NLM_F_MULTI)) {
	       if (nlh->nlmsg_len < NLMSG_ALIGN(sizeof(struct nlmsgerr))) {
	           errno = EBADMSG;
	           return 1;
	       }
	       errno = -(*((int *)NLMSG_DATA(nlh)));
	       return 1;
	   }
	   return 0;
	*/
}

// Round the length of a netlink route attribute up to align it
// properly.
func rtaAlignOf(attrlen int) int {
	return (attrlen + syscall.RTA_ALIGNTO - 1) & ^(syscall.RTA_ALIGNTO - 1)
}
