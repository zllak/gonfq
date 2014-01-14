// Package nfq provides a simple go interface to the netfilter queue library
// This is a wrapper around the C library.
package nfq

/*
 #include <stdio.h>

 #include <netinet/in.h>
 #include <linux/netfilter.h>
 #include <libnetfilter_queue/libnetfilter_queue.h>
 #cgo LDFLAGS: -lnetfilter_queue

 static u_int32_t print_pkt (struct nfq_data *tb)
 {
         int id = 0;
         struct nfqnl_msg_packet_hdr *ph;
         struct nfqnl_msg_packet_hw *hwph;
         u_int32_t mark,ifi;
         int ret;
         char *data;

         ph = nfq_get_msg_packet_hdr(tb);
         if (ph) {
                 id = ntohl(ph->packet_id);
                 printf("hw_protocol=0x%04x hook=%u id=%u ",
                         ntohs(ph->hw_protocol), ph->hook, id);
         }

         hwph = nfq_get_packet_hw(tb);
         if (hwph) {
                 int i, hlen = ntohs(hwph->hw_addrlen);

                 printf("hw_src_addr=");
                 for (i = 0; i < hlen-1; i++)
                         printf("%02x:", hwph->hw_addr[i]);
                 printf("%02x ", hwph->hw_addr[hlen-1]);
         }

         mark = nfq_get_nfmark(tb);
         if (mark)
                 printf("mark=%u ", mark);

         ifi = nfq_get_indev(tb);
         if (ifi)
                 printf("indev=%u ", ifi);
         ifi = nfq_get_outdev(tb);
         if (ifi)
                 printf("outdev=%u ", ifi);
         ifi = nfq_get_physindev(tb);
         if (ifi)
                 printf("physindev=%u ", ifi);

         ifi = nfq_get_physoutdev(tb);
         if (ifi)
                 printf("physoutdev=%u ", ifi);

         ret = nfq_get_payload(tb, &data);
         if (ret >= 0)
                 printf("payload_len=%d ", ret);

         fputc('\n', stdout);

         return id;
 }


 static int callback(struct nfq_q_handle *qh, struct nfgenmsg *nfmsg,
         struct nfq_data *nfa, void *data)
  {
      printf("entering callback\n");
      u_int32_t id = print_pkt(nfa);
      return (nfq_set_verdict(qh, id, NF_ACCEPT, 0, NULL));
  }

  // Create the queue, giving the C callback to avoid awkward cast in Go
  static struct nfq_q_handle *create_queue(struct nfq_handle *q) {
      return (nfq_create_queue(q, 0, &callback, NULL));
  }

*/
import "C"

import (
	"fmt"
	"unsafe"
)

/*
int queueHandle_input ( struct nfq_q_handle *qh, struct nfgenmsg *nfmsg, struct nfq_data *nfad, void *mdata ){

struct iphdr *ip;
int id;
struct nfqnl_msg_packet_hdr *ph = nfq_get_msg_packet_hdr ( ( struct nfq_data * ) nfad );
if ( ph ) id = ntohl ( ph->packet_id );
nfq_get_payload ( ( struct nfq_data * ) nfad, (char**)&ip );
*/

var (
	nfqOpenFailed        = fmt.Errorf("nfq: open failed")
	nfqCreateQueueFailed = fmt.Errorf("nfq: failed to create queue")
	nfqSetModeFailed     = fmt.Errorf("nfq: failed to set copy packet mode")
)

type Queue struct {
	q  *C.struct_nfq_handle
	qh *C.struct_nfq_q_handle
}

var queue *Queue

func Init() (err error) {
	fmt.Println("new Queue")
	queue = &Queue{}
	if queue.q = C.nfq_open(); queue.q == nil {
		return nfqOpenFailed
	}
	if queue.qh = C.create_queue(queue.q); queue.qh == nil {
		return nfqCreateQueueFailed
	}
	if C.nfq_set_mode(queue.qh, C.NFQNL_COPY_PACKET, 0xFFFF) < 0 {
		return nfqSetModeFailed
	}
	return
}

func Close() error {
	if queue != nil {
		fmt.Println("closing queue")
		C.nfq_destroy_queue(queue.qh)
		C.nfq_close(queue.q)
	}
	return nil
}

func ReadPackets() error {
	buf := make([]byte, 4096)

	fd := C.nfq_fd(queue.q)
	fmt.Println("fd", fd)
	for {
		fmt.Println("before recv")
		rv := C.recv(fd, unsafe.Pointer(&buf[0]), (C.size_t)(len(buf)), C.int(0))
		fmt.Println("after recv")
		if rv <= 0 {
			break
		}
		fmt.Println("read", rv, "chars")
		// Trigger packet handling
		C.nfq_handle_packet(queue.q, (*C.char)(unsafe.Pointer(&buf[0])), C.int(rv))
	}
	return nil
}
