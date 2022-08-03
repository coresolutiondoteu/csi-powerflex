# some arguments that must be supplied
ARG GOPROXY
ARG GOVERSION
ARG BASEIMAGE
ARG DIGEST

# Stage to build the driver
FROM golang:${GOVERSION} as builder
ARG GOPROXY
RUN mkdir -p /go/src
COPY ./ /go/src/
WORKDIR /go/src/
RUN CGO_ENABLED=0 \
    make build

# Stage to build the driver image
FROM $BASEIMAGE@${DIGEST} AS final
# install necessary packages
# alphabetical order for easier maintenance

# Fix for packages not found.
RUN echo $'[rhel-8-baseos]\nname=Red Hat Enterprise Linux 8 (BaseOS) - $basearch\nbaseurl=http://hb.us.dell.com/pub/redhat/RHEL8/stable/BaseOS/x86_64/os/\nenabled=1\ngpgcheck=1\ngpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-redhat-release,file:///etc/pki/rpm-gpg/RPM-GPG-KEY-redhat-beta\nskip_if_unavailable=1\n\n[rhel-8-appstream]\nname=Red Hat Enterprise Linux 8 (AppStream) - $basearch\nbaseurl=http://hb.us.dell.com/pub/redhat/RHEL8/stable/AppStream/x86_64/os/\nenabled=1\ngpgcheck=1\ngpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-redhat-release,file:///etc/pki/rpm-gpg/RPM-GPG-KEY-redhat-beta\nskip_if_unavailable=1' > /etc/yum.repos.d/dell-rpm.repo

RUN microdnf update -y && \
    microdnf install -y  \
        e4fsprogs \
        kmod \
        libaio \
        numactl \
        xfsprogs && \
    microdnf clean all
ENTRYPOINT ["/csi-vxflexos.sh"]
# copy in the driver
COPY --from=builder /go/src/csi-vxflexos /
COPY "csi-vxflexos.sh" /
RUN chmod +x /csi-vxflexos.sh
LABEL vendor="Dell Inc." \
    name="csi-powerflex" \
    summary="CSI Driver for Dell EMC PowerFlex" \
    description="CSI Driver for provisioning persistent storage from Dell EMC PowerFlex" \
    version="2.3.0" \
    license="Apache-2.0"
COPY ./licenses /licenses



